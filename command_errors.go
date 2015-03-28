// Copyright 2013 Eric Myhre
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
