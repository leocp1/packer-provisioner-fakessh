// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

// +build windows

package envmap

import (
	"strings"
)

// A directory search path
type Path struct {
	S []string
}

func (p *Path) String() string {
	return strings.Join(p.S, ";")
}

func (p *Path) UnmarshalString(s string) error {
	p.S = strings.FieldsFunc(s, func(c rune) bool { return c == ';' })
	return nil
}
