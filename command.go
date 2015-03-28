// Copyright 2013 Eric Myhre
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package gosh

import (
	"fmt"
	"os/exec"
	"sync"
	"sync/atomic"
	"syscall"
	"time"
)

const (
	/**
	 * 'Unstarted' is the state of a command that has been constructed, but execution has not yet begun.
	 */
	UNSTARTED int32 = iota

	/**
	 * 'Running' is the state of a command that has begun execution, but not yet finished.
	 */
	RUNNING

	/**
	 * 'Finished' is the state of a command that has finished normally.
	 *
	 * The exit code may or may not have been success, but at the very least we
	 * successfully observed that exit code.
	 */
	FINISHED

	/**
	 * 'Panicked' is the state of a command that at some point began execution, but has encountered
	 * serious problems.
	 *
	 * It may not be clear whether or not the command is still running, since a panic implies we no
	 * longer have completely clear visibility to the command on the underlying system.  The exit
	 * code may not be reliably known.
	 */
	PANICKED
)

func NewRunningCommand(cmd *exec.Cmd) *RunningCommand {
	return &RunningCommand{
		cmd:      cmd,
		state:    UNSTARTED,
		exitCh:   make(chan bool),
		exitCode: -1,
	}
}

type RunningCommand struct {
	mutex sync.Mutex

	/** Always access this with functions from the atomic package, and when
	* transitioning states set the status after all other fields are mutated,
	* so that checks of State() serve as a memory barrier for all. */
	state int32

	cmd *exec.Cmd
	//TODO: or: callback func() int

	/** If this is set, game over. */
	err error

	/** Wait for this to close in order to wait for the process to return. */
	exitCh chan bool

	/** Exit code if we're state==FINISHED and exit codes are possible on this platform, or
	 * -1 if we're not there yet.  Will not change after exitCh has closed. */
	exitCode int

	/** Functions to call back when the command has exited. */
	exitListeners []func(*RunningCommand)
}

func (cmd *RunningCommand) State() int32 {
	return atomic.LoadInt32(&cmd.state)
}

/** Returns true if the command is current running. */
func (cmd *RunningCommand) IsRunning() bool {
	state := cmd.State()
	return state == RUNNING
}

/** Returns true if the command has ever been started (including if the command is already finished). */
func (cmd *RunningCommand) IsStarted() bool {
	state := cmd.State()
	return state == RUNNING || state == FINISHED || state == PANICKED
}

/** Returns true if the command is finished (either gracefully, or with internal errors). */
func (cmd *RunningCommand) IsDone() bool {
	state := cmd.State()
	return state == FINISHED || state == PANICKED
}

/** Returns true if the command is finished gracefully.  (A nonzero exit code may still be set.) */
func (cmd *RunningCommand) IsFinishedGracefully() bool {
	state := cmd.State()
	return state == FINISHED
}

func (cmd *RunningCommand) Start() *RunningCommand {
	if err := cmd.startCalmly(); err != nil {
		panic(err)
	}
	return cmd
}

func (cmd *RunningCommand) startCalmly() error {
	cmd.mutex.Lock()
	defer cmd.mutex.Unlock()

	if cmd.IsStarted() {
		return nil
	}

	atomic.StoreInt32(&cmd.state, RUNNING)
	if err := cmd.cmd.Start(); err != nil {
		cmd.finalState(CommandStartError{cause: err})
		return cmd.err
	}

	go cmd.waitAndHandleExit()
	return nil
}

func (cmd *RunningCommand) waitAndHandleExit() {
	exitCode := -1
	var err error
	for err == nil && exitCode == -1 {
		exitCode, err = cmd.waitTry()
	}

	// Do one last Wait for good ol' times sake.  And to use the Cmd.closeDescriptors feature.
	cmd.cmd.Wait()

	cmd.mutex.Lock()
	defer cmd.mutex.Unlock()

	cmd.exitCode = exitCode
	cmd.finalState(err)
}

func (cmd *RunningCommand) waitTry() (int, error) {
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
	processState, err := cmd.cmd.Process.Wait()
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

func (cmd *RunningCommand) finalState(err error) {
	// must hold cmd.mutex before calling this
	// golang is an epic troll: claims to be best buddy for concurrent code, SYNC PACKAGE DOES NOT HAVE REENTRANT LOCKS
	if cmd.IsRunning() {
		if err == nil {
			atomic.StoreInt32(&cmd.state, FINISHED)
		} else {
			cmd.err = err
			atomic.StoreInt32(&cmd.state, PANICKED)
		}
		// iterate over exit listeners
		for _, cb := range cmd.exitListeners {
			func() {
				defer recover()
				cb(cmd)
			}()
		}
	}
	close(cmd.exitCh)
}

/*
	Returns the pid of the process, or -1 if it isn't started yet.
*/
func (cmd *RunningCommand) Pid() int {
	if cmd.IsStarted() {
		return cmd.cmd.Process.Pid
	} else {
		return -1
	}
}

/**
 * Add a function to be called when this command completes.
 *
 * These listener functions will be invoked after the exit code and other command
 * state is final, but before other Wait() methods unblock.
 * (This means if you want for example to log a message that a process exited, and
 * your main function is Wait()'ing for that process... if you use AddExitListener()
 * to invoke your log function then you will always get the log.)
 *
 * The listener function should complete quickly and not try to perform other blocking
 * operations or locks, since other actions are waiting until the listeners have all
 * been called.  Panics that escape the function will be silently discarded; do not
 * panic in a listener.
 *
 * If the command is already in the state FINISHED or PANICKED, the callback function
 * will be invoked immediately in the current goroutine.
 */
func (cmd *RunningCommand) AddExitListener(callback func(*RunningCommand)) {
	cmd.mutex.Lock()
	defer cmd.mutex.Unlock()

	if cmd.IsDone() {
		func() {
			defer recover()
			callback(cmd)
		}()
	} else {
		cmd.exitListeners = append(cmd.exitListeners, callback)
	}
}

/**
 * Returns a channel that will be open until the command is complete.
 * This is suitable for use in a select block.
 */
func (cmd *RunningCommand) GetExitChannel() <-chan bool {
	return cmd.exitCh
}

/**
 * Waits for the command to exit before returning.
 *
 * There are no consequences to waiting on a single command repeatedly;
 * all wait calls will return normally when the command completes.  The order
 * in which multiple wait calls will return is undefined.  Similarly, there
 * are no consequences to waiting on a command that has not yet started;
 * the function will still wait without error until the command finishes.
 * (Much friendlier than os.exec.Cmd.Wait(), neh?)
 */
func (cmd *RunningCommand) Wait() {
	<-cmd.GetExitChannel()
}

/**
 * Waits for the command to exit before returning, or for the specified duration.
 * Returns true if the return was due to the command finishing, or false if the
 * return was due to timeout.
 */
func (cmd *RunningCommand) WaitSoon(d time.Duration) bool {
	select {
	case <-time.After(d):
		return false
	case <-cmd.GetExitChannel():
		return true
	}
}

/**
 * Waits for the command to exit if it has not already, then returns the exit code.
 */
func (cmd *RunningCommand) GetExitCode() int {
	if !cmd.IsDone() {
		cmd.Wait()
	}
	return cmd.exitCode
}

/**
 * Waits for the command to exit if it has not already, or for the specified duration,
 * then either returns the exit code, or -1 if the duration expired and the command
 * still hasn't returned.
 */
func (cmd *RunningCommand) GetExitCodeSoon(d time.Duration) int {
	if cmd.WaitSoon(d) {
		return cmd.exitCode
	} else {
		return -1
	}
}
