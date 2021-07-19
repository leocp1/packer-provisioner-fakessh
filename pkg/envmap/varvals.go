// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package envmap

type String struct {
	S string
}

func (s *String) String() string {
	return s.S
}

func (s *String) UnmarshalString(ns string) error {
	s.S = ns
	return nil
}
