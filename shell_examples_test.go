package gosh_test

import (
	"fmt"

	. "github.com/polydawn/gosh"
)

func ExampleNormalFlow() {
	Sh("echo", "hello world!")

	// Output:
	// hello world!
}

func ExampleStartCollectFlow() {
	proc := Gosh("echo", "hello world!").Start()
	proc.Wait()

	// Output:
	// hello world!
}

func ExampleCollectOutput() {
	str := Gosh("echo", "hello world!").Output()
	fmt.Println(str)

	// Output:
	// hello world!
}
