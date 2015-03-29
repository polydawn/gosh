package gosh

import (
	"os"
	"strings"
)

func joinStringSlice(x, y []string) []string {
	w := len(x)
	z := make([]string, w+len(y))
	copy(z, x)
	copy(z[w:], y)
	return z
}

func getOsEnv() map[string]string {
	env := make(map[string]string)
	for _, line := range os.Environ() {
		chunks := strings.SplitN(line, "=", 2)
		env[chunks[0]] = chunks[1]
	}
	return env
}
