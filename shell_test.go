package gosh

import (
	"bytes"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestMergineCommandTemplates(t *testing.T) {
	Convey("Merging command templates", t, func() {
		one := CommandTemplate{
			Args: []string{"one"},
			Env: Env{
				"one": "one",
			},
		}
		two := CommandTemplate{
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

func TestExecIntegration(t *testing.T) {
	Convey("Given a command template", t, func() {
		cmd := CommandTemplate{
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
		cmd := CommandTemplate{
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
