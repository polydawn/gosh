package gosh

import (
	"os"
)

func main() {
	echo := Sh("echo")("-n", "-e")(DefaultIO)
	echo("wat\n", "\t\033[0;31mred and indented\033[0m\n")()

	// basic commands
	Sh("head")("--bytes=20", "/dev/zero")()

	// making your own shorthand
	aptget := Sh("apt-get")("install")
	aptyes := aptget("-y")
	// now `aptyes("git")` will install git and the `-y` argument means it will proceed automatically
	_ = aptyes

	// we'll do shorthand for a shell, so we can use the shell's echo for the rest of the demo.
	shell := Sh("bash")("-c")

	// output rules
	shell("echo 'default output is quiet'")()    // gosh defaults to NullIO.
	shell(DefaultIO)("echo 'such a message!'")() // *now* your termnial gets it!
	shell = shell(DefaultIO)                     // let's always use DefaultIO now :)

	// environment rules
	shell("echo path=$PATH")()                           // your default env is retained
	shell(ClearEnv{})("echo path=$PATH")()               // unless you don't want that, of course
	shell(Env{"VAR": "59"})("echo some_var=$SOME_VAR")() // you can easily set an env var for just this command
	shell = shell(Env{"VAR": "59"})                      // or make any env var part of the command template (try it with $PATH -- no more bashrc needed every again!)

	// directing input and output
	cat := Sh("cat")
	catIn := cat.BakeArgs("-")
	catIn(Opts{In: "piping in things is easy!\n", Out: os.Stdout})()

	// collecting output

	// channelling output

	// changing the working directory

	// failed command automatically gets you the heck outta here

	// accepting exit codes
	shell(Env{"VAR": "59"})(Opts{OkExit: []int{59}})("exit $VAR")()
}
