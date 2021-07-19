// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

// +build windows

package envmap_test

import (
	"reflect"
	"testing"

	. "github.com/leocp1/packer-provisioner-fakessh/pkg/envmap"
)

func TestPath(t *testing.T) {
	em := NewEnvMap()
	em.M["PATH"] = &Path{}

	em.AddSlice([]string{
		`PATH=C:\Users\user\AppData\Local\Microsoft\WindowsApps;`,
	})

	v, ok := em.M["PATH"]
	if !ok {
		t.Fatalf("PATH value not in envMap.M")
	}
	p, ok := v.(*Path)
	if !ok {
		t.Fatalf("PATH value not coercible to envMap.Path")
	}
	p.S = append(p.S, `C:\Users\user\custompath`)

	got := em.EnvSlice()
	expected := []string{
		`PATH=C:\Users\user\AppData\Local\Microsoft\WindowsApps;C:\Users\user\custompath`,
	}
	if !reflect.DeepEqual(got, expected) {
		t.Errorf("%#v.EnvSlice() = %#v instead of %#v", em, got, expected)
	}
}
