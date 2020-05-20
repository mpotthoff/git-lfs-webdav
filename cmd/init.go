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
	"fmt"
	"net/url"
	"os"

	"github.com/mpotthoff/git-lfs-webdav/internal"
)

// Init executes the init command
func Init(args []string) error {
	// Set the executable path of the transfer agent to our current call path
	err := internal.GitConfigSet("lfs.customtransfer.webdav.path", os.Args[0])
	if err != nil {
		return err
	}

	// Set the executable argument to "transfer"
	err = internal.GitConfigSet("lfs.customtransfer.webdav.args", "transfer")
	if err != nil {
		return err
	}

	// Force Git LFS to use us as the transfer agent without a proper LFS API
	err = internal.GitConfigSet("lfs.standalonetransferagent", "webdav")
	if err != nil {
		return err
	}

	if len(args) > 0 {
		u, err := url.Parse(args[0])
		if err != nil {
			return fmt.Errorf("Failed to parse url %q: %v", args[0], err)
		}

		// Change http/https to webdav/webdavs so that Git will definitely fail
		// on new clones. Otherwise it will try to contact the LFS URL with the
		// LFS Bulk API and might report other errors.
		if u.Scheme == "http" {
			u.Scheme = "webdav"
		} else if u.Scheme == "https" {
			u.Scheme = "webdavs"
		}

		lfsURL := u.String()

		// Save the URL inside .lfsconfig so that it can be committed to the repository.
		// Otherwise everyone that clones the repository would have to know the URL.
		err = internal.GitConfigFileSet(".lfsconfig", "lfs.url", lfsURL)
		if err != nil {
			return err
		}

		fmt.Printf("Successfully initialized LFS WebDAV with url %q!\n", lfsURL)
	} else {
		fmt.Println("Successfully initialized LFS WebDAV!")
		fmt.Println("If you have just cloned the repository run 'git reset --hard master' to checkout the LFS files.")
	}

	return nil
}
