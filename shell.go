package gosh

func Gosh(args ...interface{}) Command {
	var cmdt Opts
	cmdt.Launcher = ExecLauncher
	cmdt.Env = getOsEnv()
	cmdt.OkExit = []int{0}
	return enclose(bake(cmdt, args...))
}

/*
	Calling a `Command` merges in the arguments and then
	immediately launches a `Proc`.  The `Command()` call waits for the `Proc`
	to complete, panics if it fails, and finally returns the `Proc`.

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

type magic struct{ cmdt Opts }

func enclose(cmdt Opts) Command {
	return func(args ...interface{}) Proc {
		if len(args) == 0 {
			p := cmdt.Launcher(cmdt)
			p.Wait()
			return p
		}
		switch magic := args[0].(type) {
		case *magic:
			magic.cmdt = cmdt
			return nil
		default:
			return cmdt.Launcher(bake(cmdt, args...))
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
