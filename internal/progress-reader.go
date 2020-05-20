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
	"io"
)

// ProgressFunc is the progress callback function.
type ProgressFunc func(total int64, sinceLast int64)

// ProgressReader wraps an existing io.Reader.
type ProgressReader struct {
	io.Reader
	ProgressFunc
	total int64
}

// Read 'overrides' the underlying io.Reader's Read method.
func (pt *ProgressReader) Read(p []byte) (int, error) {
	n, err := pt.Reader.Read(p)

	sinceLast := int64(n)
	pt.total += sinceLast

	pt.ProgressFunc(pt.total, sinceLast)

	return n, err
}
