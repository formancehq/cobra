// Copyright 2013-2023 The Cobra Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cobra

import (
	"strconv"
	"strings"

	flag "github.com/spf13/pflag"
)

// boolNegValue is a pflag.Value that negates and forwards Set calls
// to the original boolean flag's Value. This is used to implement
// --no-<flag> counterparts for boolean flags.
type boolNegValue struct {
	original *flag.Flag
}

func (v *boolNegValue) Set(s string) error {
	b, err := strconv.ParseBool(s)
	if err != nil {
		return err
	}
	return v.original.Value.Set(strconv.FormatBool(!b))
}

func (v *boolNegValue) Type() string { return "bool" }

func (v *boolNegValue) String() string {
	b, err := strconv.ParseBool(v.original.Value.String())
	if err != nil {
		return "true"
	}
	return strconv.FormatBool(!b)
}

// IsBoolFlag implements pflag's boolFlag interface so that
// --no-<flag> doesn't require an explicit value argument.
func (v *boolNegValue) IsBoolFlag() bool { return true }

// initNoBoolFlags creates --no-<name> counterparts for boolean flags
// that have a default value of "true". The --no- flag, when set,
// will set the original flag's value to false.
//
// This is only called when the command (or a parent) has set
// AddNegativeBoolFlags to true.
func (c *Command) initNoBoolFlags() {
	if !c.addNegativeBoolFlags() {
		return
	}

	c.mergePersistentFlags()

	var toAdd []struct {
		name  string
		usage string
		value *boolNegValue
	}

	c.Flags().VisitAll(func(f *flag.Flag) {
		// Only consider boolean flags with default "true"
		if f.Value.Type() != "bool" || f.DefValue != "true" {
			return
		}
		// Skip flags already starting with "no-"
		if strings.HasPrefix(f.Name, "no-") {
			return
		}
		noName := "no-" + f.Name
		// Skip if a --no- flag already exists
		if c.Flags().Lookup(noName) != nil {
			return
		}
		toAdd = append(toAdd, struct {
			name  string
			usage string
			value *boolNegValue
		}{
			name:  noName,
			usage: "Negate the --" + f.Name + " flag",
			value: &boolNegValue{original: f},
		})
	})

	for _, entry := range toAdd {
		noFlag := c.Flags().VarPF(entry.value, entry.name, "", entry.usage)
		noFlag.NoOptDefVal = "true"
		// Extract the original flag name from "no-<name>"
		origName := strings.TrimPrefix(entry.name, "no-")
		c.MarkFlagsMutuallyExclusive(origName, entry.name)
	}
}

// addNegativeBoolFlags checks whether this command or any of its parents
// has requested automatic --no- flag generation for boolean flags.
func (c *Command) addNegativeBoolFlags() bool {
	for cmd := c; cmd != nil; cmd = cmd.parent {
		if cmd.AddNegativeBoolFlags {
			return true
		}
	}
	return false
}
