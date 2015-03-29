package gosh

import (
	"os"
)

/*
	Bake this in to connect all the std in/out/err
	streams to the parent process.

	This resembles the usual behavior of your interactive shell
	(but is probably not what you want if using gosh in the middle
	of a large application).
*/
var DefaultIO = Opts{
	In:  os.Stdin,
	Out: os.Stdout,
	Err: os.Stderr,
}

/*
	Bake this in to silence all output and error messages
	and disconnect the input stream.
*/
var NullIO = Opts{
	In:  nil,
	Out: nil,
	Err: nil,
}
