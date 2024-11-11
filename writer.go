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
	stdio "io"
	"time"
)

type nopDeadlineWriter struct {
	stdio.Writer
}

func NopDeadlineWriter(w stdio.Writer) DeadlineWriter {
	return nopDeadlineWriter{w}
}

func (nopDeadlineWriter) SetWriteDeadline(t time.Time) error { return nil }
