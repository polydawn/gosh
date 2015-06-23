package gosh

import (
	"fmt"
	"reflect"
	"strings"
)

/*
	`gosh.Error` is a grouping interface for all errors raised by gosh.

	Errors fall into two main headings...

	Configuration errors:
	  - IncomprehensibleCommandModifierError
	  - NoArgumentsError

	Execution errors:
	  - NoSuchCommandError
	  - ProcMonitorError
	  - FailureExitCode

	Gosh typically raises errors with panics.  This is a deliberate design
	choice to make the easiest, tersest usages of gosh feel as much as possible
	like writing a shell script with "-e" mode (exit immediately on error) set.

	All gosh errors can be distingushed by use of type switches.
	(Gosh does not use "value" type errors (i.e. `var SomeError = fmt.Errorf[...]`)
	because these are ineffective at holding information for programatic use later.)

	Gosh errors may contain special fields with additional information -- one
	near-ubiquitous example being the "cause" error -- which are always exported,
	so that you can access them after casting the error to its specific type.
	At no point should your application ever be required to parse strings in
	order to handle gosh errors.

	(Note: The authors of gosh recommend checking out a hierarchical error system,
	like the one provided by `github.com/spacemonkeygo/errors`,
	but we have not used it here in the interest of keeping gosh's
	dependencies standard-library-only.)
*/
type Error interface {
	error
	GoshError() // marker method
}

// bulk type assertion
var _ []Error = []Error{
	NoSuchCommandError{},
	NoArgumentsError{},
	NoSuchCwdError{},
	ProcMonitorError{},
	IncomprehensibleCommandModifierError{},
	FailureExitCode{},
}

/*
	NoSuchCommandError is raised when a command name (the first argument)
	cannot be found.
*/
type NoSuchCommandError struct {
	Name  string
	Cause error
}

func (err NoSuchCommandError) Error() string {
	return fmt.Sprintf("gosh: command not found: %q", err.Name)
}
func (err NoSuchCommandError) GoshError() {}

/*
	NoArgumentsErr is raised when a command template is launched but
	has no arguments.
*/
type NoArgumentsError struct{}

func (err NoArgumentsError) Error() string {
	return "gosh: no arguments specified"
}
func (err NoArgumentsError) GoshError() {}

/*
	NoSuchCwdError is raised when a command template is launched but
	has no arguments.
*/
type NoSuchCwdError struct {
	Path  string // attempted cwd path
	Cause error  // is an `*os.PathError`; may clarify whether not a dir or perm denied
}

func (err NoSuchCwdError) Error() string {
	return fmt.Sprintf("gosh: cannot use %q for cwd: %s", err.Path, err.Cause)
}
func (err NoSuchCwdError) GoshError() {}

/*
	ProcMonitorError is raised to report any errors encountered while trying
	to wait for completion, shuttle I/O, or get information about the exit
	status of a command.

	When a ProcMonitorError occurs, the Proc `Status()` will also become
	`PANICKED`, and gosh may no longer be able to reliably detect the
	state of the command.
*/
type ProcMonitorError struct {
	Cause error
}

func (err ProcMonitorError) Error() string {
	return fmt.Sprintf("gosh: error monitoring proc: %s", err.Cause)
}
func (err ProcMonitorError) GoshError() {}

/*
	Error when any of the command templating functions is called with
	arguments of an unexpected type.  The `interface{}` arguments to command
	templating functions may only be of the following types:
		- Opts
		- Env
		- ClearEnv
		- string
		- []string

	This should mostly be a compile-time problem as long as you write your
	script to not actually pass unchecked types of interface{}.
*/
type IncomprehensibleCommandModifierError struct {
	wat *interface{}
}

func (err IncomprehensibleCommandModifierError) Error() string {
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
func (err IncomprehensibleCommandModifierError) GoshError() {}

/*
	Error for commands run by Sh that exited with a non-successful status.

	What exactly qualifies as an unsuccessful status can be defined per command,
	but by default is any exit code other than zero.
*/
type FailureExitCode struct {
	Cmdname string
	Code    int
	Message string
}

func (err FailureExitCode) Error() string {
	msg := ""
	if err.Message != "" {
		msg = "\n\tCommand output was:\n\t\t\"\"\"\n\t\t" + strings.Replace(err.Message, "\n", "\n\t\t", -1) + "\n\t\t\"\"\""
	}
	return fmt.Sprintf("gosh: command \"%s\" exited with unexpected status %d%s", err.Cmdname, err.Code, msg)
}
func (err FailureExitCode) GoshError() {}
