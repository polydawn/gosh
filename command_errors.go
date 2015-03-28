package gosh

import (
	"fmt"
)

/**
 * Error encountered while trying to set up or start executing a command.
 */
type CommandStartError struct {
	cause error
}

func (err CommandStartError) Cause() error {
	return err.cause
}

func (err CommandStartError) Error() string {
	return fmt.Sprintf("error starting command: %s", err.Cause())
}

/**
 * Error encountered while trying to wait for completion, or get information about
 * the exit status of a command.
 */
type CommandMonitorError struct {
	cause error
}

func (err CommandMonitorError) Cause() error {
	return err.cause
}

func (err CommandMonitorError) Error() string {
	return fmt.Sprintf("error monitoring command: %s", err.Cause())
}
