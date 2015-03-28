package gosh

import (
	"bytes"
	"github.com/coocood/assrt"
	"os/exec"
	. "strconv"
	"sync"
	"syscall"
	"testing"
	"time"
)

func TestStateConsts(t *testing.T) {
	if !(UNSTARTED == 0 && RUNNING == 1 && FINISHED == 2 && PANICKED == 3) {
		t.Fail()
	}
}

// Test that we can exec something, wait, and it returns quickly and with an exit code.
func TestPshExecBasic(t *testing.T) {
	assert := assrt.NewAssert(t)

	cmdr := NewRunningCommand(
		exec.Command("echo"),
	)
	cmdr.startCalmly()
	cmdr.WaitSoon(1 * time.Second)
	assert.Equal(
		0,
		cmdr.GetExitCode(),
	)
	assert.Equal(
		nil,
		cmdr.err,
	)
	assert.Equal(
		FINISHED,
		cmdr.State(),
	)
}

func TestPshWaitTimeout(t *testing.T) {
	assert := assrt.NewAssert(t)

	cmdr := NewRunningCommand(
		exec.Command("sleep", "1"),
	)
	cmdr.startCalmly()
	assert.Equal(
		false,
		cmdr.WaitSoon(20*time.Millisecond),
	)
	// the go race detector would flag this as a race.
	// and correctly so!  that's why this field is private.
	// assert.Equal(
	// 	-1,
	// 	cmdr.exitCode,
	// )
	assert.Equal(
		RUNNING,
		cmdr.State(),
	)
}

func TestPshExitListeners(t *testing.T) {
	assert := assrt.NewAssert(t)

	cmdr := NewRunningCommand(
		exec.Command("echo"),
	)
	cmdr.startCalmly()
	var wg sync.WaitGroup
	wg.Add(1)
	cmdr.AddExitListener(func(*RunningCommand) {
		defer wg.Done()
		assert.Equal(
			0,
			cmdr.GetExitCode(),
		)
		assert.Equal(
			nil,
			cmdr.err,
		)
		assert.Equal(
			FINISHED,
			cmdr.State(),
		)
	})
	wg.Wait()
	assert.Equal(
		0,
		cmdr.GetExitCode(),
	)
	assert.Equal(
		nil,
		cmdr.err,
	)
	assert.Equal(
		FINISHED,
		cmdr.State(),
	)
}

func TestPshExitCode(t *testing.T) {
	assert := assrt.NewAssert(t)

	cmdr := NewRunningCommand(
		exec.Command("sh", []string{"-c", "exit 22"}...),
	)
	cmdr.startCalmly()
	cmdr.WaitSoon(1 * time.Second)
	assert.Equal(
		22,
		cmdr.GetExitCode(),
	)
	assert.Equal(
		nil,
		cmdr.err,
	)
	assert.Equal(
		FINISHED,
		cmdr.State(),
	)
}

func TestPshNonexistentCommandPanics(t *testing.T) {
	assert := assrt.NewAssert(t)

	cmdr := NewRunningCommand(
		exec.Command("/thishadbetternotbeacommand"),
	)
	cmdr.startCalmly()
	cmdr.WaitSoon(1 * time.Second)
	assert.Equal(
		-1,
		cmdr.GetExitCode(),
	)
	assert.NotEqual(
		nil,
		cmdr.err,
	)
	assert.Equal(
		PANICKED,
		cmdr.State(),
	)
}

func TestExitBySignalCodes(t *testing.T) {
	assert := assrt.NewAssert(t)

	cmdr := NewRunningCommand(
		exec.Command("sleep", "3"),
	)
	cmdr.Start()
	NewRunningCommand(exec.Command("kill", "-9", Itoa(cmdr.Pid()))).Start()
	assert.Equal(
		137,
		cmdr.GetExitCode(),
	)
	assert.Equal(
		nil,
		cmdr.err,
	)
	assert.Equal(
		FINISHED,
		cmdr.State(),
	)
}

func TestGoshDoesNotReportNondeadlySignalsAsExit(t *testing.T) {
	assert := assrt.NewAssert(t)

	cmd := exec.Command("bash", "-c",
		// this bash script does not die when it recieves a SIGINT; it catches it and exits orderly (with a different code).
		"function catch_sig () { exit 22; }; trap catch_sig 2; sleep 1; echo 'do not want reach'; exit 88;",
	)
	cmd.Stdout = &bytes.Buffer{}
	cmdr := NewRunningCommand(cmd)
	cmdr.Start()

	// Wait a moment to give the bash time to set up its trap.
	// Then spring the trap.
	time.Sleep(200 * time.Millisecond)
	NewRunningCommand(exec.Command("kill", "-2", Itoa(cmdr.Pid()))).Start()

	// There's a substantial pause before the command returns, despite the fact we killed it almost immediately.
	// Not entirely sure why.  I assume it has to do with go's concept of cleaning up before wait() returns, but I don't
	// know what it's cleaning up after -- if you play with that trap script in a regular bash, it returns immediately
	// and does not leave defunct processes around.

	assert.Equal(
		22,
		cmdr.GetExitCode(),
	)
	assert.Equal(
		nil,
		cmdr.err,
	)
	assert.Equal(
		FINISHED,
		cmdr.State(),
	)
	assert.Equal(
		"",
		cmd.Stdout.(*bytes.Buffer).String(),
	)
}

func TestGoshDoesNotReportSigStopOrContinueAsExit(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}

	assert := assrt.NewAssert(t)

	cmdr := NewRunningCommand(
		exec.Command("bash", "-c", "sleep 1; exit 4;"),
	)
	cmdr.Start()
	NewRunningCommand(exec.Command("kill", "-SIGSTOP", Itoa(cmdr.Pid()))).Start().Wait()

	// the command shouldn't be able to return while stopped, regardless of how short the sleep call is.
	assert.Equal(
		false,
		cmdr.WaitSoon(1500*time.Millisecond),
	)

	NewRunningCommand(exec.Command("kill", "-SIGCONT", Itoa(cmdr.Pid()))).Start().Wait()

	assert.Equal(
		4,
		cmdr.GetExitCode(),
	)
	assert.Equal(
		nil,
		cmdr.err,
	)
	assert.Equal(
		FINISHED,
		cmdr.State(),
	)
}

func TestGoshDoesNotReportSigStopOrContinueAsExitEvenUnderPtrace(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}

	assert := assrt.NewAssert(t)

	cmdr := NewRunningCommand(
		exec.Command("bash", "-c", "sleep 1; exit 4;"),
	)
	cmdr.Start()

	// Ride the wild wind
	if err := syscall.PtraceAttach(cmdr.Pid()); err != nil {
		panic(err)
	}

	NewRunningCommand(exec.Command("kill", "-SIGSTOP", Itoa(cmdr.Pid()))).Start().Wait()

	assert.Equal(
		false,
		cmdr.WaitSoon(1500*time.Millisecond),
	)

	NewRunningCommand(exec.Command("kill", "-SIGCONT", Itoa(cmdr.Pid()))).Start().Wait()

	// Must detach ptrace again for the wait to return.
	if err := syscall.PtraceDetach(cmdr.Pid()); err != nil {
		// This boggles my mind.  You can pause before and after this, and that pid most certainly does exist, but here we occationally get errors nonetheless.
		// Have to skip, because we are attached, that process does exist, and if we can't detach, waiting for exit is going to hang forever.
		t.Skipf("error detaching ptrace: %+v -- pid=%v\n", err, cmdr.Pid())
	}

	assert.Equal(
		4,
		cmdr.GetExitCode(),
	)
	assert.Equal(
		nil,
		cmdr.err,
	)
	assert.Equal(
		FINISHED,
		cmdr.State(),
	)
}
