package gosh

import (
	"bytes"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestMergineOptss(t *testing.T) {
	Convey("Merging command templates", t, func() {
		one := Opts{
			Args: []string{"one"},
			Env: Env{
				"one": "one",
			},
		}
		two := Opts{
			Args: []string{"two"},
			Env: Env{
				"two": "two",
			},
		}
		three := one.Merge(two)

		Convey("Args should join", func() {
			So(three.Args, ShouldResemble, []string{"one", "two"})
		})
		Convey("Args on the parents should be unchanged", FailureContinues, func() {
			So(one.Args, ShouldResemble, []string{"one"})
			So(two.Args, ShouldResemble, []string{"two"})
		})
		Convey("Env should join", func() {
			So(three.Env, ShouldResemble, Env{"one": "one", "two": "two"})
		})
		Convey("Env on the parents should be unchanged", FailureContinues, func() {
			So(one.Env, ShouldResemble, Env{"one": "one"})
			So(two.Env, ShouldResemble, Env{"two": "two"})
		})
	})
}

func TestInvocationBehaviors(t *testing.T) {
	// still presumes exec as the backing invoker, regretably
	Convey("Given a command that will succeed", t, func() {
		cmd := Opts{
			Args: []string{"echo", "success"},
		}
		Convey("RunAndReport should have no comment", func() {
			Gosh(cmd).RunAndReport()
		})
	})
	Convey("Given a command that will exit with 12", t, func() {
		cmd := Opts{
			Args: []string{"bash", "-c", "echo failuremessage 1>&2; exit 12"},
		}
		Convey("RunAndReport should panic; error should include output", func() {
			defer func() {
				err := recover()
				So(err, ShouldNotBeNil)
				So(err, ShouldHaveSameTypeAs, FailureExitCode{})
				errExit := err.(FailureExitCode)
				So(errExit.Message, ShouldEqual, "failuremessage\n")
			}()
			Gosh(cmd).RunAndReport()
		})
	})
	Convey("Given a command that will exit with 0", t, func() {
		cmd := Opts{
			Args: []string{"bash", "-c", "echo lol; echo failuremessage 1>&2"},
		}
		Convey("RunAndReport should have output", func() {
			_, output := Gosh(cmd).RunAndReport()
			So(output, ShouldResemble, "lol\nfailuremessage\n")
		})
	})
}

func TestExecIntegration(t *testing.T) {
	Convey("Given a command template", t, func() {
		cmd := Opts{
			Args: []string{"true"},
		}

		Convey("We should be able to invoke it", func() {
			p := ExecLauncher(cmd)

			Convey("It should return", func() {
				So(p.GetExitCode(), ShouldEqual, 0)
			})
		})
	})

	Convey("Given a command template with outputs", t, func() {
		var buf bytes.Buffer
		cmd := Opts{
			Args: []string{"echo", "msg"},
			Out:  &buf,
		}

		Convey("We should be able to invoke it", func() {
			p := ExecLauncher(cmd)

			Convey("It should return", func() {
				So(p.GetExitCode(), ShouldEqual, 0)
			})
			Convey("It should emit output", func() {
				p.Wait()
				So(buf.String(), ShouldEqual, "msg\n")
			})
		})
	})
}
