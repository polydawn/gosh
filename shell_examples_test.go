package gosh_test

import (
	"os"

	. "github.com/polydawn/gosh"
)

func ExampleNormalFlow() {
	cmd := Gosh("echo", "hello world!", Opts{Out: os.Stdout})
	cmd()

	// Output:
	// hello world!
}
