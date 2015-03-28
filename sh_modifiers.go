package gosh

import (
	"os"
)

type CommandTemplate struct {
	Cmd string

	Args []string

	Env Env

	Opts

	debug ShDebugListener
}

type Opts struct {
	Cwd string

	/*
		Can be a:
		  - string, in which case it will be copied in literally
		  - []byte, again, taken literally
		  - io.Reader, which will be streamed in
		  - bytes.Buffer, all that sort of thing, taken literally
		  - <-chan string, in which case that will be streamed in
		  - <-chan byte[], in which case that will be streamed in
		  - another Command, in which case that will be started with this one and its output piped into this one
	*/
	In interface{}

	/*
		Can be a:
		  - bytes.Buffer, which will be written to literally
		  - io.Writer, which will be written to streamingly, flushed to whenever the command flushes
		  - chan<- string, which will be written to streamingly, flushed to whenever a line break occurs in the output
		  - chan<- byte[], which will be written to streamingly, flushed to whenever the command flushes

		(There's nothing that's quite the equivalent of how you can give In a string, sadly; since
		strings are immutable in golang, you can't set Out=&str and get anywhere.)
	*/
	Out interface{}

	/*
		Can be all the same things Out can be, and does the same thing, but for stderr.
	*/
	Err interface{}

	/*
		Exit status codes that are to be considered "successful".  If not provided, [0] is the default.
		(If this slice is provided, zero will -not- be considered a success code unless explicitly included.)
	*/
	OkExit []int
}

var DefaultIO = Opts{
	In:  os.Stdin,
	Out: os.Stdout,
	Err: os.Stderr,
}

var NullIO = Opts{
	In:  nil,
	Out: nil,
	Err: nil,
}

type Env map[string]string

type ClearEnv struct{}
