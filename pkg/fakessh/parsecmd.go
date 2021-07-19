// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package fakessh

import (
	"fmt"
	"strings"
)

// Get the command part of ssh arguments
func ParseCmd(args []string) []string {
	flagWArg := map[string]bool{
		"-B:": true,
		"-b":  true,
		"-c":  true,
		"-D":  true,
		"-E":  true,
		"-e":  true,
		"-F":  true,
		"-I":  true,
		"-i":  true,
		"-J":  true,
		"-L":  true,
		"-l":  true,
		"-O":  true,
		"-o":  true,
		"-p":  true,
		"-Q":  true,
		"-R":  true,
		"-S":  true,
		"-W":  true,
		"-w":  true,
	}
	seenHost := false

	i := 1
	for ; i < len(args); i++ {
		if flagWArg[args[i]] {
			i++
		} else if args[i] == "--" {
			i++
			break
		} else if args[i][0] != '-' {
			if seenHost {
				break
			}
			seenHost = true
		}
	}

	return args[i:]
}

// Convert an array of strings into a sh command
//
// Currently just concatenates with a space between each argument
func ArgvToSh(cmd []string) string {
	b := strings.Builder{}
	for i, arg := range cmd {
		fmt.Fprintf(&b, "%s", arg)
		if i < len(cmd)-1 {
			fmt.Fprintf(&b, " ")
		}
	}
	return b.String()
}
