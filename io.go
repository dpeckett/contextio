// SPDX-License-Identifier: MPL-2.0
/*
 * Copyright (C) 2024 Damian Peckett <damian@pecke.tt>.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/.
 */

// Package contextio provides context-aware I/O primitives for Go.
package contextio

import (
	stdio "io"
	"time"
)

// Reader types.
type DeadlineReader interface {
	stdio.Reader
	SetReadDeadline(t time.Time) error
}

type DeadlineReadCloser interface {
	stdio.Closer
	DeadlineReader
}

// Writer types.
type DeadlineWriter interface {
	stdio.Writer
	SetWriteDeadline(t time.Time) error
}

type DeadlineWriteCloser interface {
	stdio.Closer
	DeadlineWriter
}

// Full duplex I/O.
type DeadlineReadWriter interface {
	DeadlineReader
	DeadlineWriter
}

type DeadlineReadWriteCloser interface {
	stdio.Closer
	DeadlineReadWriter
}
