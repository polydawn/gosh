package gosh

import (
	"fmt"
	"os/exec"

	"github.com/polydawn/gosh/iox"
)

var ExecLauncher Launcher = func(cmdt Opts) Proc {
	if cmdt.Args == nil || len(cmdt.Args) < 1 {
		panic(NoArgumentsErr{})
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

	// go time
	return ExecProcCmd(cmd)
}
