// Copyright (C) 2020 Creditor Corp. Group.
// See LICENSE for copying information.

package logger

// Log exposes functionality to write messages in stdout.
type Logger interface {
	// Error is used to send formatted as error message.
	Error(msg string, err error)
}
