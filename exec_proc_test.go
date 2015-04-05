package gosh

import (
	"bytes"
	"os/exec"
	"strconv"
	"sync"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

// these tests assume a variety of system commands:
// echo, sleep, sh

// do this to all test commands by default less we muck our terminal
func nilifyFDs(cmd *exec.Cmd) *exec.Cmd {
	cmd.Stdin = nil
	cmd.Stdout = nil
	cmd.Stderr = nil
	return cmd
}

func TestProcExec(t *testing.T) {
	Convey("Basic execution should work", t, FailureContinues, func() {
		cmd := nilifyFDs(exec.Command("echo"))
		p := ExecProcCmd(cmd)
		So(p.WaitSoon(1*time.Second), ShouldBeTrue)
		So(p.GetExitCode(), ShouldEqual, 0)
		So(p.State(), ShouldEqual, FINISHED)
	})

	Convey("WaitSoon should return before a slow command", t, FailureContinues, func() {
		cmd := nilifyFDs(exec.Command("sleep", "1"))
		start := time.Now()
		p := ExecProcCmd(cmd)
		So(p.WaitSoon(20*time.Millisecond), ShouldBeFalse)
		end := time.Now()
		So(p.State(), ShouldEqual, RUNNING)
		So(end, ShouldNotHappenWithin, 10*time.Millisecond, start)
		So(end, ShouldHappenWithin, 500*time.Millisecond, start)
	})

	Convey("Given an exit listener", t, func() {
		cmd := nilifyFDs(exec.Command("echo"))
		p := ExecProcCmd(cmd)
		var wg sync.WaitGroup
		wg.Add(1)

		Convey("The exit listener should be called", FailureContinues, func() {
			p.AddExitListener(func(Proc) {
				wg.Done()
			})
			wg.Wait()
			So(p.State(), ShouldEqual, FINISHED)
		})
		Convey("The exit listener should see the finished status", FailureContinues, func(c C) {
			p.AddExitListener(func(Proc) {
				defer wg.Done()
				c.So(p.State(), ShouldEqual, FINISHED)
			})
			wg.Wait()
		})
		// Note that we can't actually test that these block the Wait() return.
		// Mostly because we can't actually guarantee that at all.
		// See comments at the bottom of proc.go for discussion of the limitations.
	})

	Convey("Given a command name that cannot be found", t, func() {
		// TODO
	})

	Convey("Given commands that will exit non-zero", t, func() {
		cmd := nilifyFDs(exec.Command("sh", []string{"-c", "exit 22"}...))
		p := ExecProcCmd(cmd)
		Convey("The exit code should be reported accurately", FailureContinues, func() {
			So(p.GetExitCode(), ShouldEqual, 22)
			So(p.State(), ShouldEqual, FINISHED)
		})
	})

	Convey("Given commands that will recieve signals", t, func() {
		Convey("A process killed by a signal should exit with 128+SIG", FailureContinues, func() {
			cmd := nilifyFDs(exec.Command("sleep", []string{"3"}...))
			p := ExecProcCmd(cmd)

			// We could make this better by exposing the `exec.Cmd`, but that
			// needs a clear mechanism that doesn't ruin the Proc abstraction.
			ExecProcCmd(nilifyFDs(exec.Command("kill", "-9", strconv.Itoa(p.Pid())))).Wait()

			So(p.GetExitCode(), ShouldEqual, 137)
			So(p.State(), ShouldEqual, FINISHED)
		})
		Convey("Nondeadly signals should not be reported as exit codes", FailureContinues, func() {
			cmd := nilifyFDs(exec.Command("bash", "-c",
				// this bash script does not die when it recieves a SIGINT; it catches it and exits orderly (with a different code).
				"function catch_sig () { exit 22; }; trap catch_sig 2; sleep 1; echo 'do not want reach'; exit 88;",
			))
			var buf bytes.Buffer
			cmd.Stdout = &buf
			cmd.Stderr = &buf
			p := ExecProcCmd(cmd)

			// Wait a moment to give the bash time to set up its trap.
			// Then spring the trap.
			time.Sleep(200 * time.Millisecond)
			ExecProcCmd(nilifyFDs(exec.Command("kill", "-2", strconv.Itoa(p.Pid())))).Wait()

			// There's a substantial pause before the command returns, despite the fact we killed it almost immediately.
			// Not entirely sure why.  I assume it has to do with go's concept of cleaning up before wait() returns, but I don't
			// know what it's cleaning up after -- if you play with that trap script in a regular bash, it returns immediately
			// and does not leave defunct processes around.

			So(p.GetExitCode(), ShouldEqual, 22)
			So(p.State(), ShouldEqual, FINISHED)
			So(buf.String(), ShouldEqual, "")
		})
	})
}
