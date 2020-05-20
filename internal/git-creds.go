/**
 * Copyright (c) 2020 Michael Potthoff
 *
 * This file is part of git-lfs-webdav.
 *
 * git-lfs-webdav is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * git-lfs-webdav is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
 * GNU Lesser General Public License for more details.
 *
 * You should have received a copy of the GNU Lesser General Public License
 * along with git-lfs-webdav. If not, see <http://www.gnu.org/licenses/>.
 */

package internal

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// Creds represents a set of key/value pairs
type Creds map[string]string

func bufferCreds(creds Creds) *bytes.Buffer {
	buf := new(bytes.Buffer)

	for key, value := range creds {
		buf.Write([]byte(key))
		buf.Write([]byte("="))
		buf.Write([]byte(value))
		buf.Write([]byte("\n"))
	}

	return buf
}

func execCredential(subcommand string, input Creds) (Creds, error) {
	output := new(bytes.Buffer)

	cmd := exec.Command("git", "credential", subcommand)
	cmd.Env = append(os.Environ(), "GIT_TERMINAL_PROMPT=0")
	cmd.Stdin = bufferCreds(input)
	cmd.Stdout = output
	cmd.Stderr = os.Stderr

	err := cmd.Start()
	if err == nil {
		err = cmd.Wait()
	}

	if _, ok := err.(*exec.ExitError); ok {
		if subcommand == "fill" && err.Error() == "exit status 128" {
			return nil, nil
		}
	}

	if err != nil {
		return nil, fmt.Errorf("'git credential %s' failed with: %v", subcommand, err)
	}

	creds := make(Creds)
	for _, line := range strings.Split(output.String(), "\n") {
		pieces := strings.SplitN(line, "=", 2)
		if len(pieces) < 2 || len(pieces[1]) < 1 {
			continue
		}

		creds[pieces[0]] = pieces[1]
	}

	return creds, nil
}

// GitCredentialFill asks git to fill the given credentials
func GitCredentialFill(creds Creds) (Creds, error) {
	return execCredential("fill", creds)
}

// GitCredentialApprove marks the given credentials as approved
func GitCredentialApprove(creds Creds) error {
	_, err := execCredential("approve", creds)
	return err
}

// GitCredentialReject marks the given credentials as rejected
func GitCredentialReject(creds Creds) error {
	_, err := execCredential("reject", creds)
	return err
}
