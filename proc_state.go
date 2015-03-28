package gosh

type State int32

// we don't exactly need 32 bits, but it's often powerful to use this in a LoadInt32,
// and there's no such thing as a sub-word memory fence, so.  32 it is.

const (
	/*
		'Unstarted' is the state of a command that has been constructed, but execution has not yet begun.
	*/
	UNSTARTED State = iota

	/*
		'Running' is the state of a command that has begun execution, but not yet finished.
	*/
	RUNNING

	/*
		'Finished' is the state of a command that has finished normally.

		The exit code may or may not have been success, but at the very least we
		successfully observed that exit code.
	*/
	FINISHED

	/*
		'Panicked' is the state of a command that at some point began execution, but has encountered
		serious problems.

		It may not be clear whether or not the command is still running, since a panic implies we no
		longer have completely clear visibility to the command on the underlying system.  The exit
		code may not be reliably known.
	*/
	PANICKED
)

/*
	Returns true if the command is current running.
*/
func (state State) IsRunning() bool {
	return state == RUNNING
}

/*
	Returns true if the command has ever been started (including if the command is already finished).
*/
func (state State) IsStarted() bool {
	switch state {
	case RUNNING, FINISHED, PANICKED:
		return true
	default:
		return false
	}
}

/*
	Returns true if the command is finished (either gracefully, or with internal errors).
*/
func (state State) IsDone() bool {
	switch state {
	case FINISHED, PANICKED:
		return true
	default:
		return false
	}
}

/*
	Returns true if the command is finished gracefully.  (A nonzero exit code may still be set.)
*/
func (state State) IsFinishedGracefully() bool {
	return state == FINISHED
}
