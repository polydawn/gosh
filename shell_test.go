package gosh

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestShell(t *testing.T) {
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
