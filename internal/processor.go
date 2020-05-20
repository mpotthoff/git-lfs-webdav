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
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/studio-b12/gowebdav"
)

var (
	gitPath string
	baseURL *url.URL
	creds   Creds
	client  *gowebdav.Client
)

func createClient() {
	var username string
	var password string

	if creds != nil {
		username = creds["username"]
		password = creds["password"]
	}

	client = gowebdav.NewClient(baseURL.String(), username, password)
}

func checkAuth(err error) bool {
	// Check whether the error is of type PathError
	if perr, ok := err.(*os.PathError); ok {
		// Check whether it is an "Authorize" error and we don't already have credentials
		if perr.Op == "Authorize" && creds == nil {
			// Then ask the git credential manager

			creds = make(Creds)
			creds["url"] = baseURL.String()

			creds, _ = GitCredentialFill(creds)
			if creds != nil {
				// If we got new credentials recreate the client
				createClient()
				return true
			}
		}
	}

	return false
}

func processInit(operation string, remote string, concurrent bool, concurrentTransfer int, writer *bufio.Writer) error {
	var (
		err  error
		err2 error
	)

	gitPath, err = GitGetPath()
	if err != nil {
		return SendResponse(&InitResponse{&TransferError{1, fmt.Sprintf("Failed to get '.git' path: %v", err)}}, writer)
	}

	// First try to get the LFS URL from .git/config
	lfsURL, err := GitConfigGet("lfs.url")
	if len(lfsURL) < 1 {
		// Next try .lfsconfig
		lfsURL, err2 = GitConfigFileGet(".lfsconfig", "lfs.url")
		if len(lfsURL) < 1 {
			// Otherwise return an error
			if err != nil || err2 != nil {
				return SendResponse(&InitResponse{&TransferError{2, fmt.Sprintf("Failed to get LFS URL: %v\n%v", err, err2)}}, writer)
			}

			return SendResponse(&InitResponse{&TransferError{3, "Git LFS URL not configured!"}}, writer)
		}
	}

	baseURL, err = url.Parse(lfsURL)
	if err != nil {
		return SendResponse(&InitResponse{&TransferError{4, fmt.Sprintf("Failed to parse LFS URL %q: %v", lfsURL, err)}}, writer)
	}

	// Rewrite the URL back from webdav/webdavs to http/https
	// This was done in cmd/init
	if baseURL.Scheme == "webdav" {
		baseURL.Scheme = "http"
	} else if baseURL.Scheme == "webdavs" {
		baseURL.Scheme = "https"
	}

	// Use any credentials passed in the URL
	if baseURL.User != nil {
		username := baseURL.User.Username()
		password, passwordSet := baseURL.User.Password()
		baseURL.User = nil

		creds = make(Creds)
		creds["username"] = username
		if passwordSet {
			creds["password"] = password
		}
	} else {
		creds = nil
	}

	createClient()

	return SendResponse(&InitResponse{}, writer)
}

func processDownload(oid string, size int64, action *Action, writer *bufio.Writer) error {
	basePath := strings.ReplaceAll(filepath.Join(oid[0:2], oid[2:4]), "\\", "/")
	fullPath := strings.ReplaceAll(filepath.Join(basePath, oid), "\\", "/")

	// Try to get some information of the remote file and do some consistency checks
	remoteInfo, err := client.Stat(fullPath)
	if err != nil && checkAuth(err) {
		// If the credentials were changed retry the call
		remoteInfo, err = client.Stat(fullPath)
	}
	if err != nil {
		return SendTransferError(oid, 5, fmt.Sprintf("Failed to stat remote file %q: %v", fullPath, err), writer)
	}

	if !remoteInfo.Mode().IsRegular() {
		return SendTransferError(oid, 6, fmt.Sprintf("Remote file %q is not a regular file", fullPath), writer)
	}

	if remoteInfo.Size() != size {
		return SendTransferError(oid, 7, fmt.Sprintf("Expected size %v but got %v for remote file %q", size, remoteInfo.Size(), fullPath), writer)
	}

	// Open the remote file
	remoteReader, err := client.ReadStream(fullPath)
	if err != nil && checkAuth(err) {
		// If the credentials were changed retry the call
		remoteReader, err = client.ReadStream(fullPath)
	}
	if err != nil {
		return SendTransferError(oid, 8, fmt.Sprintf("Failed to read remote file %q: %v", fullPath, err), writer)
	}

	defer remoteReader.Close()

	// Wrap the reader in a ProgressReader which will call the given function for every Read() call to report the progress
	reader := &ProgressReader{Reader: remoteReader, ProgressFunc: func(bytesSoFar int64, bytesSinceLast int64) {
		SendProgress(oid, bytesSoFar, bytesSinceLast, writer)
	}}

	// Create a local tmp file in .git/lfs/tmp which will be the target of the download.
	// This is important so that Git LFS can later rename the file to the real destination path.
	// Otherwise this might fail in case the file was stored on a different mountpoint than the .git folder.
	tmpDir := filepath.Join(gitPath, "lfs", "tmp")
	os.MkdirAll(tmpDir, 0644)
	tmpPath := filepath.Join(tmpDir, fmt.Sprintf("%v.tmp", oid))

	// Open the local file
	file, err := os.Create(tmpPath)
	if err != nil {
		return SendTransferError(oid, 9, fmt.Sprintf("Failed to open local file %q: %v", tmpPath, err), writer)
	}

	defer file.Close()

	// Copy everything from the remote file into the local one
	_, err = io.Copy(file, reader)
	if err != nil {
		return SendTransferError(oid, 10, fmt.Sprintf("Failed to download remote file %q to local file %q: %v", fullPath, tmpPath, err), writer)
	}

	return SendResponse(&TransferResponse{Event: "complete", Oid: oid, Path: tmpPath}, writer)
}

func processUpload(oid string, size int64, action *Action, path string, writer *bufio.Writer) error {
	basePath := strings.ReplaceAll(filepath.Join(oid[0:2], oid[2:4]), "\\", "/")
	fullPath := strings.ReplaceAll(filepath.Join(basePath, oid), "\\", "/")

	// Do some consistency checks on the given information
	localInfo, err := os.Stat(path)
	if err != nil {
		return SendTransferError(oid, 11, fmt.Sprintf("Failed to stat local file %q: %v", path, err), writer)
	}

	if !localInfo.Mode().IsRegular() {
		return SendTransferError(oid, 12, fmt.Sprintf("Local file %q is not a regular file", path), writer)
	}

	if localInfo.Size() != size {
		return SendTransferError(oid, 13, fmt.Sprintf("Expected size %v but got %v for local file %q", size, localInfo.Size(), path), writer)
	}

	// Get some information about the expected remote path (to check later whether it already exists)
	remoteInfo, err := client.Stat(fullPath)
	if err != nil && checkAuth(err) {
		// If the credentials were changed retry the call
		remoteInfo, err = client.Stat(fullPath)
	}
	if err != nil {
		if strings.Contains(err.Error(), "Not Found") {
			// Ignore any Not Found errors
			remoteInfo = nil
			err = nil
		} else {
			return SendTransferError(oid, 14, fmt.Sprintf("Failed to stat remote file %q: %v", fullPath, err), writer)
		}
	}

	// Check whether the file already exists with the expected size
	if remoteInfo != nil && remoteInfo.Size() == size {
		// In that case report the upload as complete without actually uploading something

		SendProgress(oid, size, size, writer)

		return SendResponse(&TransferResponse{Event: "complete", Oid: oid}, writer)
	}

	// Create the required directory structure on the server
	err = client.MkdirAll(basePath, 0644)
	if err != nil && checkAuth(err) {
		// If the credentials were changed retry the call
		err = client.MkdirAll(basePath, 0644)
	}
	if err != nil {
		return SendTransferError(oid, 15, fmt.Sprintf("Failed to create remote folder %q: %v", basePath, err), writer)
	}

	// Open the local file
	file, err := os.Open(path)
	if err != nil {
		return SendTransferError(oid, 16, fmt.Sprintf("Failed to open local file %q: %v", path, err), writer)
	}

	defer file.Close()

	// Wrap the file in a ProgressReader which will call the given function for every Read() call to report the progress
	reader := &ProgressReader{Reader: file, ProgressFunc: func(bytesSoFar int64, bytesSinceLast int64) {
		SendProgress(oid, bytesSoFar, bytesSinceLast, writer)
	}}

	// Write the remote file
	err = client.WriteStream(fullPath, reader, 0644)
	if err != nil && checkAuth(err) {
		// If the credentials were changed retry the call
		err = client.WriteStream(fullPath, reader, 0644)
	}
	if err != nil {
		return SendTransferError(oid, 17, fmt.Sprintf("Failed to write remote file %q: %v", fullPath, err), writer)
	}

	return SendResponse(&TransferResponse{Event: "complete", Oid: oid}, writer)
}

// Processor processes the input
func Processor() error {
	scanner := bufio.NewScanner(os.Stdin)
	writer := bufio.NewWriter(os.Stdout)

	for scanner.Scan() {
		line := scanner.Text()

		var req Request
		err := json.Unmarshal([]byte(line), &req)
		if err != nil {
			return err
		}

		switch req.Event {
		case "init":
			err := processInit(req.Operation, req.Remote, req.Concurrent, req.ConcurrentTransfers, writer)
			if err != nil {
				return err
			}
		case "download":
			err := processDownload(req.Oid, req.Size, req.Action, writer)
			if err != nil {
				return err
			}
		case "upload":
			err := processUpload(req.Oid, req.Size, req.Action, req.Path, writer)
			if err != nil {
				return err
			}
		case "terminate":
			return nil
		}
	}

	return scanner.Err()
}
