package gosh

import (
	"os"
	"strings"
)

/*
	Why the golang stdlib doesn't already expose environ as a map from
	string to string -- because that's WHAT IT IS -- is so far beyond my
	understanding...
*/
func getOsEnv() map[string]string {
	env := make(map[string]string)
	for _, line := range os.Environ() {
		chunks := strings.SplitN(line, "=", 2)
		env[chunks[0]] = chunks[1]
	}
	return env
}
