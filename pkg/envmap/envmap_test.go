// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package envmap_test

import (
	"reflect"
	"testing"

	. "github.com/leocp1/packer-provisioner-fakessh/pkg/envmap"
)

func TestUnset(t *testing.T) {
	em := NewEnvMap()
	em.M["PPNC"] = &String{S: "ssh"}
	em.Unset["PPNC"] = struct{}{}

	got := em.EnvSlice()
	expected := []string{}
	if !reflect.DeepEqual(got, expected) {
		t.Errorf("%#v.EnvSlice() = %#v instead of %#v", em, got, expected)
	}
}

func TestAddSlice(t *testing.T) {
	em := NewEnvMap()
	em.M["PPNC"] = &String{S: "default_ssh"}
	em.Unset["PPNC"] = struct{}{}
	em.AddSlice([]string{
		"DONTCARE=3",
		"PPNC=ssh",
		"DONTCARE=",
	})

	got := em.EnvSlice()
	expected := []string{
		"DONTCARE=3",
		"DONTCARE=",
		"PPNC=ssh",
	}
	if !reflect.DeepEqual(got, expected) {
		t.Errorf("%#v.EnvSlice() = %#v instead of %#v", em, got, expected)
	}
}
