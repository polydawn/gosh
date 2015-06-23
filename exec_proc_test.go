package gosh

import (
	"bytes"
	"os"
	"os/exec"
	"sync"
	"syscall"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

// these tests assume a variety of system commands:
// - echo, sleep, sh
// and the existence of some usual basic linux filesystem locations:
// - /usr
// these are unfortunate presumptions, but treating them all as read-only is generally
// less problematic than creating them on the fly and thus demanding sandboxing.

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

	Convey("Gathering output from a command should work", t, FailureContinues, func() {
		cmd := nilifyFDs(exec.Command("echo", "output string"))
		var buf bytes.Buffer
		cmd.Stdout = &buf
		p := ExecProcCmd(cmd)
		So(p.GetExitCode(), ShouldEqual, 0)
		So(p.State(), ShouldEqual, FINISHED)
		So(buf.String(), ShouldEqual, "output string\n")
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
		// yes, these are different code paths.  yuck.
		// chalk this up as another thing gosh helps you with normalizing.
		Convey("Because the name can't be found", func() {
			cmd := nilifyFDs(exec.Command("surely-not-a-command"))

			Convey("Launch should fail immediately", func() {
				defer func() {
					err := recover()
					So(err, ShouldNotBeNil)
					So(err, ShouldHaveSameTypeAs, NoSuchCommandError{})
					err2 := err.(NoSuchCommandError)
					So(err2.Name, ShouldEqual, "surely-not-a-command")
				}()
				ExecProcCmd(cmd)
			})
		})
		Convey("Because part of the path doesn't exist", func() {
			cmd := nilifyFDs(exec.Command("/surely/not/a/command"))

			Convey("Launch should fail immediately", func() {
				defer func() {
					err := recover()
					So(err, ShouldNotBeNil)
					So(err, ShouldHaveSameTypeAs, NoSuchCommandError{})
					err2 := err.(NoSuchCommandError)
					So(err2.Name, ShouldEqual, "/surely/not/a/command")
				}()
				ExecProcCmd(cmd)
			})
		})
	})

	Convey("Given a working directory override", t, func() {
		cmd := nilifyFDs(exec.Command("pwd"))

		Convey("An existing directory should work", func() {
			cmd.Dir = "/usr"
			var buf bytes.Buffer
			cmd.Stdout = &buf
			p := ExecProcCmd(cmd)
			So(p.GetExitCode(), ShouldEqual, 0)
			So(p.State(), ShouldEqual, FINISHED)
			So(buf.String(), ShouldEqual, "/usr\n")
		})
		Convey("A nonexistent directory should fail immediately", func() {
			cmd.Dir = "/thisisnotapath"
			defer func() {
				err := recover()
				So(err, ShouldNotBeNil)
				So(err, ShouldHaveSameTypeAs, NoSuchCwdError{})
				err2 := err.(NoSuchCwdError)
				So(err2.Path, ShouldEqual, cmd.Dir)
			}()
			ExecProcCmd(cmd)
		})
	})

	Convey("Given a chroot", t, func() {
		if os.Getuid() != 0 {
			// well, technically merely having capability SYS_CHROOT should be enough, but this is easier to check
			SkipConvey("SKIPPED: need root", nil)
			return
		}

		cmd := nilifyFDs(exec.Command("pwd"))
		cmd.SysProcAttr = &syscall.SysProcAttr{
			Chroot: "/",
		}

		Convey("An existing directory should work", func() {
			cmd.Dir = "/usr"
			var buf bytes.Buffer
			cmd.Stdout = &buf
			p := ExecProcCmd(cmd)
			So(p.GetExitCode(), ShouldEqual, 0)
			So(p.State(), ShouldEqual, FINISHED)
			So(buf.String(), ShouldEqual, "/usr\n")
		})
		Convey("A nonexistent directory should fail immediately", func() {
			cmd.Dir = "/thisisnotapath"
			defer func() {
				err := recover()
				So(err, ShouldNotBeNil)
				So(err, ShouldHaveSameTypeAs, NoSuchCommandError{}) // THIS SUCKS
				// want: NoSuchCwdError;
				// can't do it: we literally get `&os.PathError{Op:"fork/exec", Path:"/bin/pwd", Err:0x2}` and there's no way to distinguish it.
				// to the consumer: sorry :'(

				// Specifically, it comes from the `if n == int(unsafe.Sizeof(err1)) { err = Errno(err1)` section in `forkExec()` in syscall/exec_unix.go, where it's handling errors read from a pipe to the child.
				// We could make the check ourselves from outside, but it's inherently racy to do so, meaning fully correct consumer code would have to be aware of this anyway.
				// You'll actually hit this any time you have a `SysProcAttr` set; there's a branch all the way up in `startProcess()` that makes this same choice... they actually went with the race, it seems.
			}()
			ExecProcCmd(cmd)
		})
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
		// TODO
	})
}
