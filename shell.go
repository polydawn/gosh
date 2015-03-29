package gosh

import (
	"bytes"
	"os"
)

/*
	Creates a new Command and immediately invokes it, returning when
	the process is complete.

	This is shorthand for `Gosh(args).Run()`.
*/
func Sh(args ...interface{}) Proc {
	return Gosh(args...).Run()
}

/*
	Creates a new Command with the defaults for shell-like behavior.
*/
func Gosh(args ...interface{}) Command {
	return enclose(bake(Opts{
		Launcher: ExecLauncher,
		Env:      getOsEnv(),
		In:       os.Stdin,
		Out:      os.Stdout,
		Err:      os.Stderr,
		OkExit:   []int{0},
	}, args...))
}

/*
	Calling a `Command` merges in the arguments and then
	immediately launches a `Proc`.  The `Command()` call waits for the `Proc`
	to complete, panics if it fails, and finally returns the `Proc`.

	The parameters can take many forms:
	  - `string` or `[]string` types will be merged into the command args list.
	  - `Env` types will be joined with the command environment variables.
	  - `ClearEnv` will discard *all* current environment variables.
	  - `Opts` objects can do all of the above, and also
	    set the working directory,
		set the input and output streams,
		configure the "acceptable" exit codes,
		and even inject a custom Proc Launcher.

	Use `Command.Start()` to just launch and immediately return the `Proc`
	if you want to do your own job control.
*/
type Command func(args ...interface{}) Proc

/*
	Using `Bake` on a `Command` merges in the arguments, keeping them
	as a template that applies to every Proc launched from the new `Command`
	that's returned.

	The returned command can launch `Proc`s repeatedly or futher baked, just
	like the original.

	All the same types that `Command()` can accept, `Bake()` can accept too.
	They have the same effects.

	The returned command is a "deep copy" for everything except the In/Out/Err
	readers/writers -- it's completely separated; further changes do not have
	the power to mutate the original.
	(Except when the In/Out/Err references are themselves stateful;
	we have no way to deepcopy on those without knowing what their
	implementations are).
*/
func (c Command) Bake(args ...interface{}) Command {
	return enclose(bake(c.expose(), args...))
}

/*
	Starts execution of the command, and waits until completion before returning.
	If the command does not execute successfully, a panic of type FailureExitCode
	will be emitted; use `Opts.OkExit` to configure what is considered success.

	The is exactly the behavior of a no-arg invokation on an Command, i.e.
		`Gosh("echo")()`
	and
		`Gosh("echo").Run()`
	are interchangable and behave identically.

	The behavior of an invoking Command with parameters
	is the same as baking in the extra parameters and then calling Run:
		`Gosh("echo")("my", "story")`
	and
		`Gosh("echo", "my", "story").Run()`
	and
		`Gosh("echo").Bake("my", "story").Run()`
	are all interchangable and behave identically.

	Use the `Start()` method instead if you need to run a task in the
	background, or otherwise need greater control over execution.
*/
func (c Command) Run() Proc {
	return c.expose().run()
}

/*
	Starts execution of the command, and immediately returns a `Proc` that
	can be used to track execution of the command, configure exit listeners,
	etc.
*/
func (c Command) Start() Proc {
	return c.expose().start()
}

/*
	Starts execution of the command, waits until completion, and then returns the
	accumulated output of the command as a string.  As with `Run()`, a panic will be
	emitted if the command does not execute successfully.

	This does not include output from stderr; use `CombinedOutput()` for that.

	This is shorthand equivalent to `Bake(Opts{Out:val}).Run()`; that is, it will
	overrule any previously configured output, and also it has no effect on where
	stderr will go.
*/
func (c Command) Output() string {
	var buf bytes.Buffer
	c.Bake(Opts{Out: &buf}).Run()
	return buf.String()
}

/*
	Same as `Output()`, but acts on both stdout and stderr.
*/
func (c Command) CombinedOutput() string {
	var buf bytes.Buffer
	c.Bake(Opts{Out: &buf, Err: &buf}).Run()
	return buf.String()
}

type Opts struct {
	Args []string

	Env Env

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

	/*
		The `Launcher` to use when spawning a process from this template.

		You can replace this with your own function in order to do last minute
		tweaks or logging!
	*/
	Launcher Launcher
}

// Apply 'y' to 'x', returning a new structure.  'y' trumps.
func (x Opts) Merge(y Opts) Opts {
	x.Args = joinStringSlice(x.Args, y.Args)
	x.Env = x.Env.Merge(y.Env)
	if y.Cwd != "" {
		x.Cwd = y.Cwd
	}
	if y.In != nil {
		x.In = y.In
	}
	if y.Out != nil {
		x.Out = y.Out
	}
	if y.Err != nil {
		x.Err = y.Err
	}
	if y.OkExit != nil {
		x.OkExit = y.OkExit
	}
	if y.Launcher != nil {
		x.Launcher = y.Launcher
	}
	return x
}

func (cmdt Opts) start() Proc {
	return cmdt.Launcher(cmdt)
}

func (cmdt Opts) run() Proc {
	p := cmdt.start()
	p.Wait()
	exitCode := p.GetExitCode()
	for _, okcode := range cmdt.OkExit {
		if exitCode == okcode {
			return p
		}
	}
	panic(FailureExitCode{cmdname: cmdt.Args[0], code: exitCode})
}

type magic struct{ cmdt Opts }

func enclose(cmdt Opts) Command {
	// This is the actual implementation of `Command`.
	return func(args ...interface{}) Proc {
		if len(args) == 0 {
			return cmdt.run()
		}
		switch magic := args[0].(type) {
		case *magic:
			magic.cmdt = cmdt
			return nil
		default:
			return bake(cmdt, args...).start()
		}
	}
}

func (c Command) expose() Opts {
	m := &magic{}
	c(m)
	return m.cmdt
}

func bake(cmdt Opts, args ...interface{}) Opts {
	for _, arg := range args {
		switch arg := arg.(type) {
		case Opts:
			cmdt = cmdt.Merge(arg)
		case Env:
			cmdt = cmdt.Merge(Opts{Env: arg})
		case ClearEnv:
			cmdt.Env = nil
		case string:
			cmdt = cmdt.Merge(Opts{Args: []string{arg}})
		case []string:
			cmdt = cmdt.Merge(Opts{Args: arg})
		default:
			panic(IncomprehensibleCommandModifier{wat: &arg})
		}
	}
	return cmdt
}

type Env map[string]string

type ClearEnv struct{}

func (x Env) Merge(y Env) Env {
	z := make(map[string]string, len(x)+len(y))
	for k, v := range x {
		z[k] = v
	}
	for k, v := range y {
		if v == "" {
			delete(z, k)
		} else {
			z[k] = v
		}
	}
	return z
}

func (x Env) ToSlice() []string {
	z := make([]string, len(x))
	i := 0
	for k, v := range x {
		z[i] = k + "=" + v
		i++
	}
	return z
}
