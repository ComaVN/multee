// Copyright 2023 Roel Harbers.
// Use of this source code is governed by the BEER-WARE license
// that can be found in the LICENSE file.

package multee

import "errors"

var (
	ErrClosed = errors.New("multeeReader already closed")
)
