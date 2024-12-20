// SPDX-License-Identifier: MPL-2.0
/*
 * Copyright (C) 2024 Damian Peckett <damian@pecke.tt>.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/.
 */

package contextio

import (
	"context"
	"errors"
	"fmt"
	stdio "io"
	"math"
	"net"
	"os"
	"strings"
	"syscall"
	"time"
)

const (
	bufferSize   = 4096
	pollInterval = 10 * time.Millisecond
)

// CopyContext is equivalent to `io.Copy` but with context cancellation support
// (for deadline reader/writers).
// The optional `readTimeout` parameter can be used to set a timeout for
// individual read operations, if not provided read operations will block
// indefinitely (until the context is cancelled).
func CopyContext(ctx context.Context, dst DeadlineWriter, src DeadlineReader, readTimeout *time.Duration) (written int64, err error) {
	data := make([]byte, bufferSize)

	var readTimeoutOrForever time.Duration = math.MaxInt64
	if readTimeout != nil {
		readTimeoutOrForever = *readTimeout
	}

	readTimer := time.NewTimer(readTimeoutOrForever)
	defer readTimer.Stop()

	for {
		select {
		case <-ctx.Done():
			return written, ctx.Err()
		case <-readTimer.C:
			return written, os.ErrDeadlineExceeded
		default:
		}

		if err := src.SetReadDeadline(time.Now().Add(pollInterval)); err != nil {
			if isClosed(err) {
				break
			}

			return written, fmt.Errorf("failed to set read deadline: %w", err)
		}

		nr, readErr := src.Read(data)
		if readErr != nil {
			if isClosed(readErr) {
				break
			}

			if os.IsTimeout(readErr) {
				continue
			}

			return written, readErr
		}

		readTimer.Reset(readTimeoutOrForever)

		for offset := 0; offset < nr; {
			select {
			case <-ctx.Done():
				return written, ctx.Err()
			default:
			}

			if err := dst.SetWriteDeadline(time.Now().Add(pollInterval)); err != nil {
				return written, fmt.Errorf("failed to set write deadline: %w", err)
			}

			nw, writeErr := dst.Write(data[offset:nr])
			if writeErr != nil {
				if isClosed(writeErr) {
					return written, writeErr
				}

				if os.IsTimeout(writeErr) {
					offset += nw
					if offset >= nr {
						break
					}
					continue
				}

				return written, writeErr
			}

			written += int64(nw)
			offset += nw
		}
	}

	return written, nil
}

func isClosed(err error) bool {
	if errors.Is(err, stdio.EOF) ||
		errors.Is(err, os.ErrClosed) ||
		errors.Is(err, net.ErrClosed) ||
		errors.Is(err, stdio.ErrClosedPipe) ||
		errors.Is(err, syscall.EIO) ||
		// poll.ErrFileClosing is not exposed by the poll package.
		(err != nil && strings.Contains(err.Error(), "use of closed file")) {
		return true
	}

	return false
}
