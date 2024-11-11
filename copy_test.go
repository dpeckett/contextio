// SPDX-License-Identifier: MPL-2.0
/*
 * Copyright (C) 2024 Damian Peckett <damian@pecke.tt>.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/.
 */

package contextio_test

import (
	"bytes"
	"context"
	"os"
	"testing"
	"time"

	"github.com/dpeckett/contextio"
	"github.com/stretchr/testify/require"
)

func TestCopyContext(t *testing.T) {
	t.Run("Complete", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		t.Cleanup(cancel)

		pr, pw := contextio.Pipe()

		go func() {
			defer pw.Close()

			_, _ = pw.Write([]byte("hello world"))
		}()

		var dst bytes.Buffer
		n, err := contextio.CopyContext(ctx, contextio.NopDeadlineWriter(&dst), pr, nil)
		require.NoError(t, err)

		require.Equal(t, int64(11), n)
		require.Equal(t, "hello world", dst.String())
	})

	t.Run("Cancelled", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())

		pr, pw := contextio.Pipe()
		t.Cleanup(func() {
			_ = pr.Close()
			_ = pw.Close()
		})

		go func() {
			// Twice the configured poll interval.
			time.Sleep(200 * time.Millisecond)

			cancel()
		}()

		var dst bytes.Buffer
		n, err := contextio.CopyContext(ctx, contextio.NopDeadlineWriter(&dst), pr, nil)

		require.ErrorIs(t, err, context.Canceled)
		require.Zero(t, n)
	})

	t.Run("Read Timeout", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		t.Cleanup(cancel)

		pr, pw := contextio.Pipe()
		t.Cleanup(func() {
			_ = pr.Close()
			_ = pw.Close()
		})

		readTimeout := 500 * time.Millisecond

		var dst bytes.Buffer
		n, err := contextio.CopyContext(ctx, contextio.NopDeadlineWriter(&dst), pr, &readTimeout)

		require.ErrorIs(t, err, os.ErrDeadlineExceeded)
		require.Zero(t, n)
	})
}
