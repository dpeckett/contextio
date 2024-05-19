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
	"bytes"
	"context"
	"testing"
	"time"

	"github.com/noisysockets/contextio"
	"github.com/stretchr/testify/assert"
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
		n, err := contextio.CopyContext(ctx, contextio.NopDeadlineWriter(&dst), pr)
		require.NoError(t, err)

		assert.Equal(t, int64(11), n)
		assert.Equal(t, "hello world", dst.String())
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
		n, err := contextio.CopyContext(ctx, contextio.NopDeadlineWriter(&dst), pr)

		assert.ErrorIs(t, err, context.Canceled)
		assert.Zero(t, n)
	})
}
