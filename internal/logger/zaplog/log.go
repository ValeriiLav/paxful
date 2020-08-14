// Copyright (C) 2020 Creditor Corp. Group.
// See LICENSE for copying information.

package zaplog

import (
	"go.uber.org/zap"

	"paxful/internal/logger"
)

// ensures that zaplog implements logger.Logger.
var _ logger.Logger = (*zaplog)(nil)

// zaplog is an implementation of logger.Logger using zap.
type zaplog struct {
	client *zap.Logger
}

// NewLog is a constructor for a logger.Logger.
func NewLog() logger.Logger {
	return &zaplog{
		client: zap.NewExample(),
	}
}

// Error is used to send formatted as error message.
func (log zaplog) Error(msg string, err error) {
	log.client.Error(msg, zap.Error(err))
}
