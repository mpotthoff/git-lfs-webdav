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
	"path/filepath"
	"strings"
)

// GitGetPath gets the path of the '.git' directory
func GitGetPath() (string, error) {
	output := new(bytes.Buffer)

	cmd := exec.Command("git", "rev-parse", "--git-dir")
	cmd.Env = append(os.Environ(), "GIT_TERMINAL_PROMPT=0")
	cmd.Stdout = output
	cmd.Stderr = os.Stderr

	err := cmd.Start()
	if err == nil {
		err = cmd.Wait()
	}

	if err != nil {
		return "", fmt.Errorf("'git rev-parse --git-dir' failed with: %v", err)
	}

	path := strings.TrimSpace(output.String())

	path, err = filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("Failed to get absolute path of %s: %v", path, err)
	}

	path, err = filepath.EvalSymlinks(path)
	if err != nil {
		return "", fmt.Errorf("Failed to evaluate symlinks of %s: %v", path, err)
	}

	return path, nil
}

// GitConfigGet executes 'git config --get <name>'
func GitConfigGet(name string) (string, error) {
	output := new(bytes.Buffer)

	cmd := exec.Command("git", "config", "--get", name)
	cmd.Env = append(os.Environ(), "GIT_TERMINAL_PROMPT=0")
	cmd.Stdout = output
	cmd.Stderr = os.Stderr

	err := cmd.Start()
	if err == nil {
		err = cmd.Wait()
	}

	if err != nil {
		return "", fmt.Errorf("'git config --get %s' failed with: %v", name, err)
	}

	return strings.TrimSpace(output.String()), nil
}

// GitConfigFileGet executes 'git config -f <file> --get <name>'
func GitConfigFileGet(file string, name string) (string, error) {
	output := new(bytes.Buffer)

	cmd := exec.Command("git", "config", "-f", file, "--get", name)
	cmd.Env = append(os.Environ(), "GIT_TERMINAL_PROMPT=0")
	cmd.Stdout = output
	cmd.Stderr = os.Stderr

	err := cmd.Start()
	if err == nil {
		err = cmd.Wait()
	}

	if err != nil {
		return "", fmt.Errorf("'git config -f %s --get %s' failed with: %v", file, name, err)
	}

	return strings.TrimSpace(output.String()), nil
}

// GitConfigSet executes 'git config <name> <value>'
func GitConfigSet(name string, value string) error {
	cmd := exec.Command("git", "config", name, value)
	cmd.Env = append(os.Environ(), "GIT_TERMINAL_PROMPT=0")
	cmd.Stderr = os.Stderr

	err := cmd.Start()
	if err == nil {
		err = cmd.Wait()
	}

	if err != nil {
		return fmt.Errorf("'git config %s %q' failed with %v", name, value, err)
	}

	return nil
}

// GitConfigFileSet executes 'git config -f <file> <name> <value>'
func GitConfigFileSet(file string, name string, value string) error {
	cmd := exec.Command("git", "config", "-f", file, name, value)
	cmd.Env = append(os.Environ(), "GIT_TERMINAL_PROMPT=0")
	cmd.Stderr = os.Stderr

	err := cmd.Start()
	if err == nil {
		err = cmd.Wait()
	}

	if err != nil {
		return fmt.Errorf("'git config -f %s %s %q' failed with %v", file, name, value, err)
	}

	return nil
}

// GitConfigUnset executes 'git config --unset <name>'
func GitConfigUnset(name string) error {
	cmd := exec.Command("git", "config", "--unset", name)
	cmd.Env = append(os.Environ(), "GIT_TERMINAL_PROMPT=0")
	cmd.Stderr = os.Stderr

	err := cmd.Start()
	if err == nil {
		err = cmd.Wait()
	}

	if err != nil {
		return fmt.Errorf("'git config --unset %s' failed with %v", name, err)
	}

	return nil
}

// GitConfigFileUnset executes 'git config -f <file> --unset <name>'
func GitConfigFileUnset(file string, name string) error {
	cmd := exec.Command("git", "config", "-f", file, "--unset", name)
	cmd.Env = append(os.Environ(), "GIT_TERMINAL_PROMPT=0")
	cmd.Stderr = os.Stderr

	err := cmd.Start()
	if err == nil {
		err = cmd.Wait()
	}

	if err != nil {
		return fmt.Errorf("'git config -f %s --unset %s' failed with %v", file, name, err)
	}

	return nil
}
