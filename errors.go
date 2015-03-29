package gosh

import (
	"fmt"
	"reflect"
)

/*
	Error encountered while trying to set up or start executing a command.
*/
type ProcStartError struct {
	cause error
}

func (err ProcStartError) Cause() error {
	return err.cause
}

func (err ProcStartError) Error() string {
	return fmt.Sprintf("error starting proc: %s", err.Cause())
}

var NoArgumentsErr = fmt.Errorf("no arguments specified")

/*
	Error encountered while trying to wait for completion, or get information about
	the exit status of a command.
*/
type ProcMonitorError struct {
	cause error
}

func (err ProcMonitorError) Cause() error {
	return err.cause
}

func (err ProcMonitorError) Error() string {
	return fmt.Sprintf("error monitoring proc: %s", err.Cause())
}

/*
	Error when Sh() or its family of functions is called with arguments of an unexpected
	type.  Sh() functions only expect arguments of the public types declared in the
	sh_modifiers.go file when setting up a command.

	This should mostly be a compile-time problem as long as you write your
	script to not actually pass unchecked types of interface{} into Sh() commands.
*/
type IncomprehensibleCommandModifier struct {
	wat *interface{}
}

func (err IncomprehensibleCommandModifier) Error() string {
	return fmt.Sprintf("gosh: incomprehensible command modifier: do not want type \"%v\"", whoru(reflect.ValueOf(*err.wat)))
}

func whoru(val reflect.Value) string {
	kind := val.Kind()
	typ := val.Type()

	if kind == reflect.Ptr {
		return fmt.Sprintf("*%s", whoru(val.Elem()))
	} else if kind == reflect.Interface {
		return whoru(val.Elem())
	} else {
		return typ.Name()
	}
}

/*
	Error for commands run by Sh that exited with a non-successful status.

	What exactly qualifies as an unsuccessful status can be defined per command,
	but by default is any exit code other than zero.
*/
type FailureExitCode struct {
	Cmdname string
	Code    int
}

func (err FailureExitCode) Error() string {
	return fmt.Sprintf("gosh: command \"%s\" exited with unexpected status %d", err.Cmdname, err.Code)
}
