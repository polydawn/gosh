package gosh

import (
	"bytes"
	"fmt"
	"os/exec"

	"github.com/polydawn/gosh/iox"
)

func Sh(cmd string) Command {
	var cmdt CommandTemplate
	cmdt.Cmd = cmd
	cmdt.Env = getOsEnv()
	cmdt.OkExit = []int{0}
	return enclose(&cmdt)
}

type Command func(args ...interface{}) Command

type ShDebugListener func(cmdt *CommandTemplate)

// private type, used exactly once to create a const nobody else can create so we can use it as a flag to trigger private behavior
type expose_t bool

const expose expose_t = true

type exposer struct{ cmdt *CommandTemplate }

func closure(cmdt CommandTemplate, args ...interface{}) Command {
	if len(args) == 0 {
		// an empty call is a synonym for Command.Run().
		// if you want to just get a RunningCommand reference to track, use Command.Start() instead.
		enclose(&cmdt).Run()
		return nil
	} else if args[0] == expose {
		// produce a function that when called with an exposer, exposes its cmdt.
		return func(x ...interface{}) Command {
			t := x[0].(*exposer)
			t.cmdt = &cmdt
			return nil
		}
	} else {
		// examine each of the arguments, modify our (already forked) cmdt, and
		//  return a new callable Command closure with the newly baked command template.
		for _, rarg := range args {
			switch arg := rarg.(type) {
			case string:
				cmdt.bakeArgs(arg)
			case Env:
				cmdt.bakeEnv(arg)
			case ClearEnv:
				cmdt.clearEnv()
			case Opts:
				cmdt.bakeOpts(arg)
			default:
				panic(IncomprehensibleCommandModifier{wat: &rarg})
			}
		}
		return enclose(&cmdt)
	}
}

func (f Command) expose() *CommandTemplate {
	var t exposer
	f(expose)(&t)
	return t.cmdt
}

func enclose(cmdt *CommandTemplate) Command {
	return func(x ...interface{}) Command {
		return closure(*cmdt, x...)
	}
}

func (f Command) BakeArgs(args ...string) Command {
	return enclose(f.expose().bakeArgs(args...))
}

func (cmdt *CommandTemplate) bakeArgs(args ...string) *CommandTemplate {
	cmdt.Args = append(cmdt.Args, args...)
	return cmdt
}

func (f Command) BakeEnv(args Env) Command {
	return enclose(f.expose().bakeEnv(args))
}

func (cmdt *CommandTemplate) bakeEnv(args Env) *CommandTemplate {
	//FIXME: fork the map
	for k, v := range args {
		if v == "" {
			delete(cmdt.Env, k)
		} else {
			cmdt.Env[k] = v
		}
	}
	return cmdt
}

func (f Command) ClearEnv() Command {
	return enclose(f.expose().clearEnv())
}

func (cmdt *CommandTemplate) clearEnv() *CommandTemplate {
	cmdt.Env = make(map[string]string)
	return cmdt
}

func (f Command) BakeOpts(args ...Opts) Command {
	return enclose(f.expose().bakeOpts(args...))
}

func (cmdt *CommandTemplate) bakeOpts(args ...Opts) *CommandTemplate {
	for _, arg := range args {
		if arg.Cwd != "" {
			cmdt.Cwd = arg.Cwd
		}
		if arg.In != nil {
			cmdt.In = arg.In
		}
		if arg.Out != nil {
			cmdt.Out = arg.Out
		}
		if arg.Err != nil {
			cmdt.Err = arg.Err
		}
		if arg.OkExit != nil {
			cmdt.OkExit = arg.OkExit
		}
	}
	return cmdt
}

func (f Command) Debug(cb ShDebugListener) Command {
	return enclose(f.expose().bakeDebug(cb))
}

func (cmdt *CommandTemplate) bakeDebug(cb ShDebugListener) *CommandTemplate {
	cmdt.debug = cb
	return cmdt
}

/*
	Starts execution of the command.  Returns a reference to a RunningCommand,
	which can be used to track execution of the command, configure exit listeners,
	etc.
*/
func (f Command) Start() *RunningCommand {
	cmdt := f.expose()

	if cmdt.debug != nil {
		cmdt.debug(cmdt)
	}

	rcmd := exec.Command(cmdt.Cmd, cmdt.Args...)

	// set up env
	if cmdt.Env != nil {
		rcmd.Env = make([]string, len(cmdt.Env))
		i := 0
		for k, v := range cmdt.Env {
			rcmd.Env[i] = fmt.Sprintf("%s=%s", k, v)
			i++
		}
	}

	// set up opts (cwd/stdin/stdout/stderr)
	if cmdt.Cwd != "" {
		rcmd.Dir = cmdt.Cwd
	}
	if cmdt.In != nil {
		switch in := cmdt.In.(type) {
		case Command:
			//TODO something marvelous
			panic(fmt.Errorf("not yet implemented"))
		default:
			rcmd.Stdin = iox.ReaderFromInterface(in)
		}
	}
	if cmdt.Out != nil {
		rcmd.Stdout = iox.WriterFromInterface(cmdt.Out)
	}
	if cmdt.Err != nil {
		if cmdt.Err == cmdt.Out {
			rcmd.Stderr = rcmd.Stdout
		} else {
			rcmd.Stderr = iox.WriterFromInterface(cmdt.Err)
		}
	}

	// go time
	cmd := NewRunningCommand(rcmd)
	cmd.Start()
	return cmd
}

/*
	Starts execution of the command, and waits until completion before returning.
	If the command does not execute successfully, a panic of type FailureExitCode
	will be emitted; use Opts.OkExit to configure what is considered success.

	The is exactly the behavior of a no-arg invokation on an Command, i.e.
		`Sh("echo")()`
	and
		`Sh("echo").Run()`
	are interchangable and behave identically.

	Use the Start() method instead if you need to run a task in the background, or
	if you otherwise need greater control over execution.
*/
func (f Command) Run() {
	cmdt := f.expose()
	cmd := f.Start()
	cmd.Wait()
	exitCode := cmd.GetExitCode()
	for _, okcode := range cmdt.OkExit {
		if exitCode == okcode {
			return
		}
	}
	panic(FailureExitCode{cmdname: cmdt.Cmd, code: exitCode})
}

/*
	Starts execution of the command, waits until completion, and then returns the
	accumulated output of the command as a string.  As with Run(), a panic will be
	emitted if the command does not execute successfully.

	This does not include output from stderr; use CombinedOutput() for that.

	This acts as BakeOpts() with a value set on the Out field; that is, it will
	overrule any previously configured output, and also it has no effect on where
	stderr will go.
*/
func (f Command) Output() string {
	var buf bytes.Buffer
	f.BakeOpts(Opts{Out: &buf}).Run()
	return buf.String()
}

/*
	Same as Output(), but acts on both stdout and stderr.
*/
func (f Command) CombinedOutput() string {
	var buf bytes.Buffer
	f.BakeOpts(Opts{Out: &buf, Err: &buf}).Run()
	return buf.String()
}
