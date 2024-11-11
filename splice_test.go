// SPDX-License-Identifier: MPL-2.0
/*
 * Copyright (C) 2024 The Noisy Sockets Authors.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/.
 */

package contextio_test

import (
	"context"
	"io"
	"os"
	"testing"
	"time"

	"github.com/dpeckett/contextio"
	"github.com/stretchr/testify/require"
)

func TestSpliceContext(t *testing.T) {
	t.Run("Complete", func(t *testing.T) {
		rwa := &deadlineReadWriter{
			readBuf: []byte("hello world"),
		}

		rwb := &deadlineReadWriter{
			readBuf: []byte("dlrow olleh"),
		}

		ctx := context.Background()
		n, err := contextio.SpliceContext(ctx, rwa, rwb, nil)
		require.NoError(t, err)

		require.Equal(t, int64(22), n)

		require.Equal(t, "dlrow olleh", string(rwa.writeBuf))
		require.Equal(t, "hello world", string(rwb.writeBuf))
	})

	t.Run("Read Timeout", func(t *testing.T) {
		rwa := &nopDeadlineReadWriter{}
		rwb := &nopDeadlineReadWriter{}

		readTimeout := 500 * time.Millisecond

		ctx := context.Background()
		n, err := contextio.SpliceContext(ctx, rwa, rwb, &readTimeout)

		require.ErrorIs(t, err, os.ErrDeadlineExceeded)
		require.Zero(t, n)
	})

}

type deadlineReadWriter struct {
	readBuf  []byte
	readPos  int
	writeBuf []byte
}

func (rw *deadlineReadWriter) Read(p []byte) (n int, err error) {
	n = copy(p, rw.readBuf[rw.readPos:])
	if n == 0 {
		return 0, io.EOF
	}
	rw.readPos += n
	return n, nil
}

func (rw *deadlineReadWriter) Write(p []byte) (n int, err error) {
	rw.writeBuf = append(rw.writeBuf, p...)
	return len(p), nil
}

func (rw *deadlineReadWriter) SetReadDeadline(t time.Time) error { return nil }

func (rw *deadlineReadWriter) SetWriteDeadline(t time.Time) error { return nil }

type nopDeadlineReadWriter struct {
}

func (rw *nopDeadlineReadWriter) Read(p []byte) (n int, err error) {
	time.Sleep(10 * time.Millisecond)
	return 0, os.ErrDeadlineExceeded
}

func (rw *nopDeadlineReadWriter) Write(p []byte) (n int, err error) {
	time.Sleep(10 * time.Millisecond)
	return 0, os.ErrDeadlineExceeded
}

func (rw *nopDeadlineReadWriter) SetReadDeadline(t time.Time) error { return nil }

func (rw *nopDeadlineReadWriter) SetWriteDeadline(t time.Time) error { return nil }
