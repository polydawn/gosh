package gosh

import (
	"os"
	"time"
)

type Launcher func(Opts) Proc

/*
	Proc observes and manipulates a running command.

	It is similar to `exec.Cmd`, but applied to an in-flight process (setup
	is performed by a separate interface).  Unlike contracts of `exec.Cmd`,
	all functions on Proc are safe to call repeatedly, and in any order.
*/
type Proc interface {
	State() State

	/*
		Returns the pid of the process, or -1 if it isn't started yet.

		If an implementation doesn't have a concept of 'pid', it should return 0.
	*/
	Pid() int

	/*
		Returns a channel that will be open until the command is complete.
		This is suitable for use in a select block.
	*/
	WaitChan() <-chan struct{}

	/*
		Waits for the command to exit before returning.

		There are no consequences to waiting on a single command repeatedly;
		all wait calls will return normally when the command completes.  The order
		in which multiple wait calls will return is undefined.  Similarly, there
		are no consequences to waiting on a command that has not yet started;
		the function will still wait without error until the command finishes.
		(Much friendlier than os.exec.Cmd.Wait(), neh?)
	*/
	Wait()

	/*
		Waits for the command to exit before returning, or for the specified duration.
		Returns true if the return was due to the command finishing, or false if the
		return was due to timeout.
	*/
	WaitSoon(d time.Duration) bool

	/*
		Waits for the command to exit if it has not already, then returns the exit code.
	*/
	GetExitCode() int

	/*
		Waits for the command to exit if it has not already, or for the specified duration,
		then either returns the exit code, or -1 if the duration expired and the command
		still hasn't returned.
	*/
	GetExitCodeSoon(d time.Duration) int

	/*
		Add a function to be called when this process completes.

		These listener functions will be invoked after the exit code and other command
		state is final, but before other `Wait()` methods unblock.
		(This means if you want for example to log a message that a process exited, and
		your main function is going to exit immediately after it's finished
		Wait()'ing for that process... if you use AddExitListener()
		to invoke your log function then you will always get the log.)

		The listener function should complete quickly and not try to perform other blocking
		operations or locks, since other actions are waiting until the listeners have all
		been called.

		Panics that escape the function may result in undefined behavior, and failure
		to call other listeners; do not panic in a listener.
		Consider sending any errors to a (buffered!!) channel instead.

		If the command is already in the state FINISHED or PANICKED, the callback function
		will be invoked immediately in the current goroutine.
	*/
	AddExitListener(callback func(Proc))

	Kill()

	Signal(os.Signal)
}

// TODO: The template system should know how to accept exit listeners up front.
// This will allow setup of exit listeners that can be guaranteed to run before other Wait methods unblock,
// which will otherwise be impossible with our stance of Procs may start to run (and
// thus may exit!) the moment they're created.
//
// Not sure exactly how to cleanly implement this.
// Maybe an unexpected procConfig interface that we're careful to not ever return?
// Does have to be exported though.  Foreign implementations of Proc are supposed to be
// possible and shouldn't have to compromise their semantics.

// TODO: Reconsider if the level of caveats on exit listeners is sane.
// There's really no reason we couldn't run every one of them in a new goroutine.
// Except that's a little dumb.  If one *does* need a long/complicated/blocking
// handler, that's the exception, not the rule; and it should fire its own goroutine.
// Well, except that denies it the Wait()-blocking behavior.  But that probably
// shouldn't be used in such a situation anyway...?
