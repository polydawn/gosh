package gosh

import (
	"fmt"
	"os/exec"
	"sync"
	"sync/atomic"
	"syscall"
	"time"
)

var _ Proc = &ExecProc{}

/*
	`gosh.Proc` implementation using `os/exec`.
*/
type ExecProc struct {
	/*
		Guards all major transitions... *including* to `state`.

		`state` is safe to access by fenced read, but all changes to it are
		under this mutex (because there's other matching checks and changes
		going on with every state transition anyway).
	*/
	mutex sync.Mutex

	/*
		Always access this with functions from the atomic package, and when
		transitioning states set the status after all other fields are mutated,
		so that checks of State() serve as a memory barrier for all.

		This is actually a `gosh.State`, but `atomic` functions don't understand
		typedefs, so we keep it as an int and coerce it whenever we expose it.
	*/
	state int32

	cmd *exec.Cmd

	/* If set, a major error (i.e. status=PANICKED; does not include nonzero exit statuses). */
	err error

	/* Wait for this to close in order to wait for the process to return. */
	exitCh chan struct{}

	/*
		Exit code if we're state==FINISHED and exit codes are possible on this platform, or
		-1 if we're not there yet.  Will not change after exitCh has closed.
	*/
	exitCode int

	/* Functions to call back when the command has exited. */
	exitListeners []func(Proc)
}

func ExecProcCmd(cmd *exec.Cmd) Proc {
	p := &ExecProc{
		cmd:      cmd,
		state:    int32(UNSTARTED),
		exitCh:   make(chan struct{}),
		exitCode: -1,
	}
	if err := p.start(); err != nil {
		panic(err)
	}
	return p
}

func (p *ExecProc) State() State {
	return State(atomic.LoadInt32(&p.state))
}

func (p *ExecProc) Pid() int {
	if p.State().IsStarted() {
		return p.cmd.Process.Pid
	} else {
		return -1
	}
}

func (p *ExecProc) WaitChan() <-chan struct{} {
	return p.exitCh
}

func (p *ExecProc) Wait() {
	<-p.WaitChan()
}

func (p *ExecProc) WaitSoon(d time.Duration) bool {
	select {
	case <-time.After(d):
		return false
	case <-p.WaitChan():
		return true
	}
}

func (p *ExecProc) GetExitCode() int {
	if !p.State().IsDone() {
		p.Wait()
	}
	return p.exitCode
}

func (p *ExecProc) GetExitCodeSoon(d time.Duration) int {
	if p.WaitSoon(d) {
		return p.exitCode
	} else {
		return -1
	}
}

func (p *ExecProc) AddExitListener(callback func(Proc)) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if p.State().IsDone() {
		// TODO: a better standard of panic handling here
		callback(p)
	} else {
		p.exitListeners = append(p.exitListeners, callback)
	}
}

//
// Below lieth Guts
//

func (p *ExecProc) start() error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if p.State().IsStarted() {
		return nil
	}

	atomic.StoreInt32(&p.state, int32(RUNNING))
	if err := p.cmd.Start(); err != nil {
		p.transitionFinal(ProcMonitorError{Cause: err})
		return p.err
	}

	go p.waitAndHandleExit()
	return nil
}

func (p *ExecProc) waitAndHandleExit() {
	exitCode := -1
	var err error
	for err == nil && exitCode == -1 {
		exitCode, err = p.waitTry()
	}

	// Do one last Wait for good ol' times sake.  And to use the Cmd.closeDescriptors feature.
	p.cmd.Wait()

	p.mutex.Lock()
	defer p.mutex.Unlock()

	p.exitCode = exitCode
	p.transitionFinal(err)
}

func (p *ExecProc) waitTry() (int, error) {
	// The docs for os.Process.Wait() state "Wait waits for the Process to exit".
	// IT LIES.
	//
	// On unixy systems, under some states, os.Process.Wait() *also* returns for signals and other state changes.  See comments below, where waitStatus is being checked.
	// To actually wait for the process to exit, you have to Wait() repeatedly and check if the system-dependent codes are representative of real exit.
	//
	// You can *not* use os/exec.Cmd.Wait() to reliably wait for a command to exit on unix.  Can.  Not.  Do it.
	// os/exec.Cmd.Wait() explicitly sets a flag to see if you've called it before, and tells you to go to hell if you have.
	// Since Cmd.Wait() uses Process.Wait(), the latter of which cannot function correctly without repeated calls, and the former of which forbids repeated calls...
	// Yep, it's literally impossible to use os/exec.Cmd.Wait() correctly on unix.
	//
	processState, err := p.cmd.Process.Wait()
	if err != nil {
		return -1, err
	}

	if waitStatus, ok := processState.Sys().(syscall.WaitStatus); ok {
		if waitStatus.Exited() {
			return waitStatus.ExitStatus(), nil
		} else if waitStatus.Signaled() {
			// In bash, when a processs ends from a signal, the $? variable is set to 128+SIG.
			// We follow that same convention here.
			// So, a process terminated by ctrl-C returns 130.  A script that died to kill-9 returns 137.
			return int(waitStatus.Signal()) + 128, nil
		} else {
			// This should be more or less unreachable.
			//  ... the operative word there being "should".  Read: "you wish".
			// WaitStatus also defines Continued and Stopped states, but in practice, they don't (typically) appear here,
			//  because deep down, syscall.Wait4 is being called with options=0, and getting those states would require
			//  syscall.Wait4 being called with WUNTRACED or WCONTINUED.
			// However, syscall.Wait4 may also return the Continued and Stoppe states if ptrace() has been attached to the child,
			//  so, really, anything is possible here.
			// And thus, we have to return a special code here that causes wait to be tried in a loop.
			return -1, nil
		}
	} else {
		panic(fmt.Errorf("gosh only works systems with posix-style process semantics."))
	}
}

func (p *ExecProc) transitionFinal(err error) {
	// must hold cmd.mutex before calling this
	// golang is an epic troll: claims to be best buddy for concurrent code, SYNC PACKAGE DOES NOT HAVE REENTRANT LOCKS
	if p.State().IsRunning() {
		if err == nil {
			atomic.StoreInt32(&p.state, int32(FINISHED))
		} else {
			p.err = err
			atomic.StoreInt32(&p.state, int32(PANICKED))
		}
		// iterate over exit listeners
		for _, cb := range p.exitListeners {
			func() {
				// TODO: a better standard of panic handling here
				cb(p)
			}()
		}
	}
	close(p.exitCh)
}
