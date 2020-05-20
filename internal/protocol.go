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
	"bufio"
	"encoding/json"
	"time"
)

// Header struct
type Header struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// Action struct
type Action struct {
	Href      string            `json:"href"`
	Header    map[string]string `json:"header,omitempty"`
	ExpiresAt time.Time         `json:"expires_at,omitempty"`
}

// TransferError generic transfer error
type TransferError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// Request generic request
type Request struct {
	Event               string  `json:"event"`
	Operation           string  `json:"operation"`
	Remote              string  `json:"remote"`
	Concurrent          bool    `json:"concurrent"`
	ConcurrentTransfers int     `json:"concurrenttransfers"`
	Oid                 string  `json:"oid"`
	Size                int64   `json:"size"`
	Path                string  `json:"path"`
	Action              *Action `json:"action"`
}

// InitResponse init response
type InitResponse struct {
	Error *TransferError `json:"error,omitempty"`
}

// TransferResponse generic transfer response
type TransferResponse struct {
	Event string         `json:"event"`
	Oid   string         `json:"oid"`
	Path  string         `json:"path,omitempty"` // always blank for upload
	Error *TransferError `json:"error,omitempty"`
}

// ProgressResponse generic transfer progress response
type ProgressResponse struct {
	Event          string `json:"event"`
	Oid            string `json:"oid"`
	BytesSoFar     int64  `json:"bytesSoFar"`
	BytesSinceLast int64  `json:"bytesSinceLast"`
}

// SendResponse sends a response to Git LFS
func SendResponse(r interface{}, writer *bufio.Writer) error {
	b, err := json.Marshal(r)
	if err != nil {
		return err
	}

	_, err = writer.Write(append(b, '\n'))
	if err != nil {
		return err
	}

	writer.Flush()
	return nil
}

// SendTransferError sends an error to Git LFS
func SendTransferError(oid string, code int, message string, writer *bufio.Writer) error {
	resp := &TransferResponse{"complete", oid, "", &TransferError{code, message}}
	return SendResponse(resp, writer)
}

// SendProgress reports progress on operations
func SendProgress(oid string, bytesSoFar int64, bytesSinceLast int64, writer *bufio.Writer) error {
	resp := &ProgressResponse{"progress", oid, bytesSoFar, bytesSinceLast}
	return SendResponse(resp, writer)
}
