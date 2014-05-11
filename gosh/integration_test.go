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
	"bytes"
	"fmt"
	"github.com/coocood/assrt"
	"testing"
)

var Printf = fmt.Printf

func TestIntegration_ShReturns(t *testing.T) {
	Sh("echo")()
}

func TestIntegration_CommandProvidesExitCode(t *testing.T) {
	assert := assrt.NewAssert(t)

	exitingShellCmd := Sh("bash")("-c")("exit 14").Start()
	assert.Equal(
		14,
		exitingShellCmd.GetExitCode(),
	)
}

func TestIntegration_ShCustomSuccessCodes(t *testing.T) {
	// panics if the exit code opt doesnt work
	Sh("bash")("-c")("exit 14")(Opts{OkExit: []int{14}})()
}

func TestIntegration_ShOutputWithStringChan(t *testing.T) {
	assert := assrt.NewAssert(t)

	out := make(chan string, 1)
	Sh("echo")("wat")(Opts{Out: out})()
	assert.Equal(
		"wat\n",
		<-out,
	)
}

func TestIntegration_ShOutputWithByteSliceChan(t *testing.T) {
	assert := assrt.NewAssert(t)

	out := make(chan []byte, 1)
	Sh("echo")("wat")(Opts{Out: out})()
	assert.Equal(
		[]byte("wat\n"),
		<-out,
	)
}

func TestIntegration_ShOutputWithBuffer(t *testing.T) {
	assert := assrt.NewAssert(t)

	var out bytes.Buffer
	// note that we set opts with &out!  it's quite critical that that be a reference.
	Sh("echo")("wat")(Opts{Out: &out})()
	assert.Equal(
		"wat\n",
		out.String(),
	)
}

func TestIntegration_ShInputWithString(t *testing.T) {
	assert := assrt.NewAssert(t)

	msg := "bees"
	out := make(chan string, 1)
	Sh("cat")("-")(Opts{In: msg, Out: out})()
	assert.Equal(
		msg,
		<-out,
	)
}

func TestIntegration_ShInputWithByteSlice(t *testing.T) {
	assert := assrt.NewAssert(t)

	msg := []byte("bees")
	out := make(chan []byte, 1)
	Sh("cat")("-")(Opts{In: msg, Out: out})()
	assert.Equal(
		msg,
		<-out,
	)
}

func TestIntegration_ShInputWithStringChan(t *testing.T) {
	assert := assrt.NewAssert(t)

	msg := "bees"
	in := make(chan string, 1)
	in <- msg
	close(in)
	out := make(chan string, 1)
	Sh("cat")("-")(Opts{In: in, Out: out})()
	assert.Equal(
		msg,
		<-out,
	)
}

func TestIntegration_ShInputWithBuffer(t *testing.T) {
	assert := assrt.NewAssert(t)

	var in bytes.Buffer
	msg := "bees"
	in.WriteString(msg)
	out := make(chan string, 1)
	Sh("cat")("-")(Opts{In: msg, Out: out})()
	assert.Equal(
		msg,
		<-out,
	)
}

func TestIntegration_ShStreamingInputAndOutputWithStringChan(t *testing.T) {
	assert := assrt.NewAssert(t)

	msg1 := "bees\n"
	msg2 := "knees\n"
	in := make(chan string, 1)
	out := make(chan string, 1)
	catCmd := Sh("cat")("-")(Opts{In: in, Out: out}).Start()

	in <- msg1
	assert.Equal(
		msg1,
		<-out,
	)
	in <- msg2
	assert.Equal(
		msg2,
		<-out,
	)
	close(in)
	assert.Equal(
		0,
		catCmd.GetExitCode(),
	)
}

func TestIntegration_ShOutput(t *testing.T) {
	assert := assrt.NewAssert(t)

	cmd := Sh("sh")("-c", "echo out ; echo err 1>&2 ;")

	assert.Equal(
		"out\n",
		cmd.Output(),
	)
}

func TestIntegration_ShCombinedOutput(t *testing.T) {
	assert := assrt.NewAssert(t)

	cmd := Sh("sh")("-c", "echo out ; echo err 1>&2 ;")

	assert.Equal(
		"out\nerr\n",
		cmd.CombinedOutput(),
	)
}

func TestIntegration_NotATty(t *testing.T) {
	assert := assrt.NewAssert(t)

	out := make(chan string, 1)
	cmd := Sh("tty")(Opts{Out: out})
	p := cmd.Start()

	assert.Equal(
		"not a tty\n",
		<- out,
	)
	assert.Equal(
		1,
		p.GetExitCode(),
	)
}

func TestIntegration_ShDebug(t *testing.T) {
	assert := assrt.NewAssert(t)

	echo := Sh("echo")("alpha", "beta")

	debugArgs := []string{"nope"}
	echo = echo.Debug(func(cmdt *CommandTemplate) {
		debugArgs = cmdt.Args
	})

	echo()

	assert.Equal(
		[]string{"alpha", "beta"},
		debugArgs,
	)
}
