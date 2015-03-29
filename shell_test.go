package gosh

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestShell(t *testing.T) {
	Convey("Merging command templates", t, func() {
		one := CommandTemplate{
			Args: []string{"one"},
		}
		two := CommandTemplate{
			Args: []string{"two"},
		}
		three := one.Fold(two)
		Convey("Args should join", func() {
			So(three.Args, ShouldResemble, []string{"one", "two"})
		})
		Convey("Args on the parents should be unchanged", FailureContinues, func() {
			So(one.Args, ShouldResemble, []string{"one"})
			So(two.Args, ShouldResemble, []string{"two"})
		})
	})
}
