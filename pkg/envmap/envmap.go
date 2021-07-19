// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

// Manipulate environment variable string slices used in `exec.Cmd.Env`,
// `os.Environ()` and others.
package envmap

import (
	"fmt"
	"strings"
)

// A type that can be converted to/from a string
type VarVal interface {
	// VarVal -> String
	fmt.Stringer
	// String -> VarVal: set VarVal to string
	UnmarshalString(string) error
}

// A map representation of an environment
type EnvMap struct {
	// Env variables to marshall/unmarshall
	M map[string]VarVal
	// Set of env variables in M that should be considered unset
	Unset map[string]struct{}
	// Env vars to add but not marshall/unmarshall
	Unparsed []string
}

func NewEnvMap() *EnvMap {
	return &EnvMap{
		M:     make(map[string]VarVal),
		Unset: make(map[string]struct{}),
	}
}

func splitSet(s string) (string, string) {
	ss := strings.SplitN(s, "=", 2)
	if len(ss) < 2 {
		return s, ""
	} else {
		return ss[0], ss[1]
	}
}

// Add variables from a slice of variable assignments like the output of
// `os.Environ()`
//
// WARNING: if UnmarshalString fails, the function will return its error
// immediately, leaving the EnvMap in a partially modified state
func (em *EnvMap) AddSlice(es []string) error {
	for _, set := range es {
		name, val := splitSet(set)
		_, ok := em.Unset[name]
		if ok {
			delete(em.Unset, name)
		}
		e, ok := em.M[name]
		if ok {
			err := e.UnmarshalString(val)
			if err != nil {
				return err
			}
		} else {
			em.Unparsed = append(em.Unparsed, set)
		}
	}
	return nil
}

// Covert EnvMap into a slice of variable assignments in a similar format to
// `exec.Cmd.Env`
func (em *EnvMap) EnvSlice() []string {
	rs := make([]string, 0, len(em.M)+len(em.Unparsed))
	for _, s := range em.Unparsed {
		name, _ := splitSet(s)
		_, ok := em.Unset[name]
		if !ok {
			rs = append(rs, s)
		}
	}
	for name, val := range em.M {
		_, ok := em.Unset[name]
		if !ok {
			rs = append(rs, fmt.Sprintf("%s=%s", name, val.String()))
		}
	}
	return rs
}
