package gosh

import (
	"fmt"
	"os/exec"

	"github.com/polydawn/gosh/iox"
)

var _ Launcher = ExecLauncher

/*
	Launches a process via stdlib `exec`.
*/
func ExecLauncher(cmdt Opts) Proc {
	return execLauncher(cmdt, nil)
}

/*
	Launches a process via stdlib `exec`, and gives the provided hook function
	a shot at running against the `exec.Cmd` to do any last-minute
	preparations -- this is useful if you need to set exotic attributes like
	`SysProcAttr` which are not normally exposed by Gosh's shellish layer.
*/
func ExecCustomizingLauncher(trailingHook func(*exec.Cmd)) Launcher {
	return func(cmdt Opts) Proc {
		return execLauncher(cmdt, trailingHook)
	}
}

func execLauncher(cmdt Opts, trailingHook func(*exec.Cmd)) Proc {
	if cmdt.Args == nil || len(cmdt.Args) < 1 {
		panic(NoArgumentsError{})
	}
	cmd := exec.Command(cmdt.Args[0], cmdt.Args[1:]...)

	// set up env
	if cmdt.Env != nil {
		cmd.Env = cmdt.Env.ToSlice()
	}
	if cmdt.Cwd != "" {
		cmd.Dir = cmdt.Cwd
	}

	// set up io (stdin/stdout/stderr)
	if cmdt.In != nil {
		switch in := cmdt.In.(type) {
		case Command:
			//TODO something marvelous
			panic(fmt.Errorf("not yet implemented"))
		default:
			cmd.Stdin = iox.ReaderFromInterface(in)
		}
	}
	if cmdt.Out != nil {
		cmd.Stdout = iox.WriterFromInterface(cmdt.Out)
	}
	if cmdt.Err != nil {
		if cmdt.Err == cmdt.Out {
			cmd.Stderr = cmd.Stdout
		} else {
			cmd.Stderr = iox.WriterFromInterface(cmdt.Err)
		}
	}

	// hook, down here it's your time
	if trailingHook != nil {
		trailingHook(cmd)
	}

	// go time
	return ExecProcCmd(cmd)
}
