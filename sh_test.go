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
	"github.com/coocood/assrt"
	"testing"
)

func TestShConstruction(t *testing.T) {
	assert := assrt.NewAssert(t)

	echo := Sh("echo")

	assert.Equal(
		"echo",
		echo.expose().Cmd,
	)
}

func TestShBakeArgs(t *testing.T) {
	assert := assrt.NewAssert(t)

	echo := Sh("echo")
	echo = echo.BakeArgs("a", "b")
	echo = echo.BakeArgs("c")

	assert.Equal(
		[]string{"a", "b", "c"},
		echo.expose().Args,
	)
}

func TestShBakeArgsMagic(t *testing.T) {
	assert := assrt.NewAssert(t)

	echo := Sh("echo")("a", "b")("c")

	assert.Equal(
		[]string{"a", "b", "c"},
		echo.expose().Args,
	)
}

func TestShBakeArgsForked(t *testing.T) {
	assert := assrt.NewAssert(t)

	echo := Sh("echo")
	echo1 := echo.BakeArgs("a", "b")
	echo2 := echo.BakeArgs("c")

	assert.Equal(
		0,
		len(echo.expose().Args),
	)
	assert.Equal(
		[]string{"a", "b"},
		echo1.expose().Args,
	)
	assert.Equal(
		[]string{"c"},
		echo2.expose().Args,
	)
}

func TestShBakeArgsMagicForked(t *testing.T) {
	assert := assrt.NewAssert(t)

	echo := Sh("echo")
	echo1 := echo("a", "b")
	echo2 := echo("c")

	assert.Equal(
		0,
		len(echo.expose().Args),
	)
	assert.Equal(
		[]string{"a", "b"},
		echo1.expose().Args,
	)
	assert.Equal(
		[]string{"c"},
		echo2.expose().Args,
	)
}

func TestShBakeArgsForkedDeeper(t *testing.T) {
	assert := assrt.NewAssert(t)

	echo := Sh("echo").BakeArgs("")
	echo1 := echo.BakeArgs("a", "b")
	echo2 := echo.BakeArgs("c")

	assert.Equal(
		[]string{""},
		echo.expose().Args,
	)
	assert.Equal(
		[]string{"", "a", "b"},
		echo1.expose().Args,
	)
	assert.Equal(
		[]string{"", "c"},
		echo2.expose().Args,
	)
}

func TestShBakeArgsMagicForkedDeeper(t *testing.T) {
	assert := assrt.NewAssert(t)

	echo := Sh("echo")("")
	echo1 := echo("a", "b")
	echo2 := echo("c")

	assert.Equal(
		[]string{""},
		echo.expose().Args,
	)
	assert.Equal(
		[]string{"", "a", "b"},
		echo1.expose().Args,
	)
	assert.Equal(
		[]string{"", "c"},
		echo2.expose().Args,
	)
}

func TestShBakeEnvForked(t *testing.T) {
	t.Skip("BROKEN")
	assert := assrt.NewAssert(t)

	echo := Sh("echo").ClearEnv()
	echo1 := echo.BakeEnv(Env{"VAR": "red"})
	echo2 := echo.BakeEnv(Env{"VAR": "blue"})

	assert.Equal(
		0,
		len(echo.expose().Env),
	)
	assert.Equal(
		"red",
		echo1.expose().Env["VAR"],
	)
	assert.Equal(
		"blue",
		echo2.expose().Env["VAR"],
	)
}
