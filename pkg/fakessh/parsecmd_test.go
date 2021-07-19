// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package fakessh_test

import (
	"fmt"
	"reflect"
	"testing"

	. "github.com/leocp1/packer-provisioner-fakessh/pkg/fakessh"
)

func TestParseCmd(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		expected []string
	}{
		{
			name:     "echo",
			input:    []string{"ssh", "user@localhost", "--", "echo", "test"},
			expected: []string{"echo", "test"},
		},
		{
			name: "nix base",
			input: []string{"ssh", "user@host", "-x", "-a",
				"-i", "$HOME/.ssh/id_rsa.pub",
				"-C",
				"nix-daemon --stdio",
			},
			expected: []string{"nix-daemon --stdio"},
		},
		{
			name: "nix socket",
			input: []string{"ssh", "user@host", "-x", "-a",
				"-S", "/tmp/ssh.sock",
				"nix-daemon --stdio",
			},
			expected: []string{"nix-daemon --stdio"},
		},
		{
			name: "nix chatty",
			input: []string{"ssh", "user@host", "-x", "-a",
				"-v",
				"nix-daemon --stdio",
			},
			expected: []string{"nix-daemon --stdio"},
		},
		{
			name: "nix socket create",
			input: []string{"ssh", "user@host", "-M", "-N",
				"-S", "/tmp/ssh.sock",
				"-o", "LocalCommand=echo started",
				"-o", "PermitLocalCommand=yes",
			},
			expected: []string{},
		},
	}
	for i, tt := range tests {
		t.Run(fmt.Sprintf("%d: %s", i, tt.name), func(t *testing.T) {
			got := ParseCmd(tt.input)
			if !reflect.DeepEqual(got, tt.expected) {
				t.Errorf(
					"failed for %#v ... (expected %#v, but got %#v)",
					tt.input,
					tt.expected,
					got,
				)
			}
		})
	}
}

func TestArgvToSh(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		expected string
	}{
		{
			name:     "bell",
			input:    []string{"printf", "\a\n"},
			expected: "printf \a\n",
		},
		{
			name:     "nix-daemon",
			input:    []string{"nix-daemon --stdio"},
			expected: "nix-daemon --stdio",
		},
		{
			name:     "empty",
			input:    []string{},
			expected: "",
		},
	}
	for i, tt := range tests {
		t.Run(fmt.Sprintf("%d: %s", i, tt.name), func(t *testing.T) {
			got := ArgvToSh(tt.input)
			if got != tt.expected {
				t.Errorf(
					"failed for %#v ... (expected %#v, but got %#v)",
					tt.input,
					tt.expected,
					got,
				)
			}
		})
	}
}
