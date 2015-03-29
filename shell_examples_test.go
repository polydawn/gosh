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

func ExampleEnvVars() {
	str := Gosh(ClearEnv{}, Env{"key": "val"}, "env").Output()
	fmt.Println(str)

	// Output:
	// key=val
}

func ExampleErrorExit() {
	defer func() {
		err := recover()
		fmt.Println("code", err.(FailureExitCode).Code)
	}()
	Sh("bash", "-c", "exit 22")

	// Output:
	// code 22
}

func ExampleOkExit() {
	Sh("bash", "-c", "exit 22", Opts{OkExit: AnyExit})
	// (no output; point is just that it doesn't panic)

	// Output:
}
