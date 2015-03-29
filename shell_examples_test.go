package gosh_test

import (
	"fmt"
	"os"

	. "github.com/polydawn/gosh"
)

func ExampleNormalFlow() {
	Gosh("echo", "hello world!", Opts{Out: os.Stdout}).Run()

	// Output:
	// hello world!
}

func ExampleStartCollectFlow() {
	proc := Gosh("echo", "hello world!", Opts{Out: os.Stdout}).Start()
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
