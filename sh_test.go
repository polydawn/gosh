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
