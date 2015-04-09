package gosh

import (
	"os/exec"
	"strconv"
	"sync"
	"syscall"
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
			signal(p.Pid(), "9")

			So(p.GetExitCode(), ShouldEqual, 137)
			So(p.State(), ShouldEqual, FINISHED)
		})
		conveyFast := func(x ...interface{}) {
			if testing.Short() {
				SkipConvey(x...)
			} else {
				Convey(x...)
			}
		}
		conveyFast("Nondeadly signals should not be reported as exit codes", FailureContinues, func() {
			cmd := nilifyFDs(exec.Command("bash", "-c",
				// this bash script does not die when it recieves a SIGINT; it catches it and exits orderly (with a different code).
				"function catch_sig () { exit 22; }; trap catch_sig 2; { sleep 2 & } ; wait ; exit 88;",
				// "function catch_sig () { echo 'trap' ; date ; jobs ; echo 'disowning' ; disown ; jobs ; exit 22; }; trap catch_sig 2; echo 'start' ; date ; { (sleep 2 ; echo 'sleep survived' ; date) & }; jobs ; echo 'prewait' ; date ; wait ; echo 'postwait' ; date ; echo 'do not want reach'; exit 88;",
			))
			p := ExecProcCmd(cmd)

			// Wait a moment to give the bash time to set up its trap.
			// Then spring the trap.
			// SLOW: it would be better if the shell could tell us when it's ready
			time.Sleep(100 * time.Millisecond)
			signal(p.Pid(), "2")

			So(p.GetExitCode(), ShouldEqual, 22)
			So(p.State(), ShouldEqual, FINISHED)
		})
		conveyFast("Stop/Cont signals should not be reported as exit codes", FailureContinues, func() {
			cmd := nilifyFDs(exec.Command("bash", "-c", "sleep 1; exit 4;"))
			p := ExecProcCmd(cmd)

			signal(p.Pid(), "SIGSTOP")
			signal(p.Pid(), "SIGCONT")

			// SLOW: this waits for the entire `sleep` process
			So(p.GetExitCode(), ShouldEqual, 4)
			So(p.State(), ShouldEqual, FINISHED)
		})
		conveyFast("Stop/Cont signals should not be reported as exit codes, even under ptrace", FailureContinues, func() {
			// This exercises that really bizzare 'else' case in `waitTry` and
			// that whole retry loop around it.
			cmd := nilifyFDs(exec.Command("bash", "-c", "sleep 1; exit 4;"))
			p := ExecProcCmd(cmd)

			// Ride the wild wind
			So(syscall.PtraceAttach(p.Pid()), ShouldBeNil)
			signal(p.Pid(), "SIGSTOP")
			signal(p.Pid(), "SIGCONT")
			So(syscall.PtraceDetach(p.Pid()), ShouldBeNil)

			// SLOW: this waits for the entire `sleep` process
			So(p.GetExitCode(), ShouldEqual, 4)
			So(p.State(), ShouldEqual, FINISHED)
		})
	})
}

func signal(pid int, sig string) {
	So(ExecProcCmd(nilifyFDs(exec.Command("kill", "-"+sig, strconv.Itoa(pid)))).GetExitCode(), ShouldEqual, 0)
}
