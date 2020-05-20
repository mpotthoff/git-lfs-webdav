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

package cmd

import (
	"bufio"
	"fmt"
	"net/url"
	"os"
	"strings"
	"syscall"

	"github.com/mpotthoff/git-lfs-webdav/internal"

	"golang.org/x/crypto/ssh/terminal"
)

// Login executes the login command
func Login(args []string) error {
	reader := bufio.NewReader(os.Stdin)

	// Always read the URL from .lfsconfig in case it changes in the repository.
	rawLFSURL, err := internal.GitConfigFileGet(".lfsconfig", "lfs.url")
	if err != nil {
		return err
	}

	lfsURL, err := url.Parse(rawLFSURL)
	if err != nil {
		return fmt.Errorf("Failed to parse url %q: %v", rawLFSURL, err)
	}

	fmt.Printf("Username for %q: ", lfsURL.Host)
	username, err := reader.ReadString('\n')
	if err != nil {
		fmt.Println("")
		return fmt.Errorf("Failed to read username: %v", err)
	}

	username = strings.TrimSpace(username)

	fmt.Printf("Password for %q: ", lfsURL.Host)
	passwordBytes, err := terminal.ReadPassword(int(syscall.Stdin))
	fmt.Println("")
	if err != nil {
		return fmt.Errorf("Failed to read password: %v", err)
	}

	password := strings.TrimSpace(string(passwordBytes))

	if len(password) > 0 {
		lfsURL.User = url.UserPassword(username, password)
	} else if len(username) > 0 {
		lfsURL.User = url.User(username)
	} else {
		lfsURL.User = nil
	}

	newLFSURL := lfsURL.String()

	if rawLFSURL != newLFSURL {
		// Save the URL with the login credentials in .git/config
		// because otherwise someone might commit it accidentally.
		err := internal.GitConfigSet("lfs.url", newLFSURL)
		if err != nil {
			return err
		}

		fmt.Println("Successfully set login credentials!")
	} else {
		// If the URL is exactly the same just remove it from .git/config
		// so that the one in .lfsconfig is used instead.
		err := internal.GitConfigUnset("lfs.url")
		if err != nil {
			return err
		}

		fmt.Println("Successfully unset login credentials.")
	}

	return nil
}
