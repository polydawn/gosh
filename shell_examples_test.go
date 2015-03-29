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

func ExamplePipeline() {
	// make sure the pipeline is big enough.. or use the next example
	pipe := make(chan string, 3)
	Sh("echo", "3\n1\n2", Opts{Out: pipe})
	close(pipe)
	Sh("sort", Opts{In: pipe})

	// Output:
	// 1
	// 2
	// 3
}

func ExamplePipelineBetter() {
	// start both tasks before waiting; this means the pipeline never chokes
	pipe := make(chan string)
	job1 := Gosh("echo", "3\n1\n2", Opts{Out: pipe}).Start()
	job2 := Gosh("sort", Opts{In: pipe}).Start()
	job1.Wait()
	close(pipe)
	job2.Wait()

	// Output:
	// 1
	// 2
	// 3
}
