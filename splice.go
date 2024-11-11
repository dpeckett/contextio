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
	"time"

	"golang.org/x/sync/errgroup"
)

// SpliceContext copies data between two ReadWriters until EOF is reached on
// one of them.
// The optional `readTimeout` parameter can be used to set a timeout for
// individual read operations, if not provided read operations will block
// indefinitely (until the context is cancelled).
func SpliceContext(ctx context.Context, rwa DeadlineReadWriter, rwb DeadlineReadWriter,
	readTimeout *time.Duration) (written int64, err error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	var g errgroup.Group

	g.Go(func() error {
		defer func() {
			// Time for the other direction to complete (if necessary).
			time.Sleep(pollInterval)
			cancel()
		}()

		n, err := CopyContext(ctx, rwa, rwb, readTimeout)
		written += n
		return err
	})

	g.Go(func() error {
		defer func() {
			// Time for the other direction to complete (if necessary).
			time.Sleep(pollInterval)
			cancel()
		}()

		n, err := CopyContext(ctx, rwb, rwa, readTimeout)
		written += n
		return err
	})

	if err := g.Wait(); err != nil && !errors.Is(err, context.Canceled) {
		return written, err
	}

	return written, nil
}
