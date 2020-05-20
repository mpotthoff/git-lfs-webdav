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

package main

import (
	"fmt"
	"os"

	"github.com/mpotthoff/git-lfs-webdav/cmd"
)

func main() {
	var command string = ""
	if len(os.Args) >= 2 {
		command = os.Args[1]
	}

	var err error = nil
	switch command {
	case "init":
		err = cmd.Init(os.Args[2:])
	case "login":
		err = cmd.Login(os.Args[2:])
	case "transfer":
		err = cmd.Transfer(os.Args[2:])
	case "version":
		err = cmd.Version(os.Args[2:])
	default:
		usage := `Usage:
    git-lfs-webdav init [url]  Initialize LFS WebDAV for the git repository in the current working directory.
    git-lfs-webdav login       Save login credentials in case the git credential manager does not work.
    git-lfs-webdav transfer    Called internally by git-lfs.
    git-lfs-webdav version     Report the version number and exit.
`

		fmt.Fprint(os.Stderr, usage)
		os.Exit(1)
	}

	if err == nil {
		os.Exit(0)
	} else {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
