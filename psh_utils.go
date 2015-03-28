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
	"os"
	"strings"
)

/**
 * Why the golang stdlib doesn't already expose environ as a map from
 * string to string -- beacuse that's WHAT IT IS -- is so far beyond my
 * understanding...
 */
func getOsEnv() map[string]string {
	env := make(map[string]string)
	for _, line := range os.Environ() {
		chunks := strings.SplitN(line, "=", 2)
		env[chunks[0]] = chunks[1]
	}
	return env
}
