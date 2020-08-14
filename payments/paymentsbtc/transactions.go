// Copyright (C) 2020 Creditor Corp. Group.
// See LICENSE for copying information.

package paymentsbtc

import (
	"context"

	"github.com/zeebo/errs"

	"paxful/internal/logger"
	"paxful/payments"
)

// ensures that transactions implements payments.Transactions.
var _ payments.Transactions = (*transactions)(nil)

// Error is an error class for internal bitcoin transaction service error.
var Error = errs.Class("bitcoin transaction error")

// Config stores needed information for eth payment service initialization.
type Config struct {
	URL           string
	PrivateKey    string
	GasLimit      uint64
	GasPriceInWei int64
}

// transactions is an BTC implementation of paxful payment service.
type transactions struct {
	log               logger.Logger
	config            Config
	commissionPercent float64
}

// NewClient is a constructor for a BTC client.
func NewTransactions(log logger.Logger, config Config, commissionPercent float64) (payments.Transactions, error) {
	return &transactions{
		log:               log,
		config:            config,
		commissionPercent: commissionPercent,
	}, nil
}

// Commit is used to send transaction to a receiver.
func (t *transactions) Commit(ctx context.Context, tx payments.Transaction) (payments.Transaction, error) {
	return payments.Transaction{}, Error.New("not implemented")
}
