package gosh

func Gosh(args ...string) Command {
	var cmdt CommandTemplate
	cmdt.Args = args
	cmdt.Env = getOsEnv()
	cmdt.OkExit = []int{0}
	return wrap(cmdt)
}

type Command func(args ...interface{}) Command

type CommandTemplate struct {
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
}

// Apply 'y' to 'x', returning a new structure.  'y' trumps.
func (x CommandTemplate) Merge(y CommandTemplate) CommandTemplate {
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
	return x
}

func wrap(cmdt CommandTemplate) Command {
	return func(args ...interface{}) Command {
		return bake(cmdt, args)
	}
}

func bake(cmdt CommandTemplate, args ...interface{}) Command {
	// TODO fill with stars
	return wrap(cmdt)
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
