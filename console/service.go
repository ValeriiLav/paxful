// Copyright (C) 2020 Creditor Corp. Group.
// See LICENSE for copying information.

package console

import (
	"context"

	"github.com/zeebo/errs"

	"paxful/payments"
)

// Error is the default paxful payment service error.
var (
	Error           = errs.Class("payment console service error")
	ValidationError = errs.Class("payment console service validation error")
)

// Service exposes all payment console related logic.
type Service struct {
	payments payments.PaymentProvider
	txDB     payments.TransactionsDB
}

// NewService is a constructor for payments console Service.
//
// architecture: Service
func NewService(provider payments.PaymentProvider, txDB payments.TransactionsDB) *Service {
	return &Service{
		payments: provider,
		txDB:     txDB,
	}
}

// CommitTx will commit transaction through payment service.
func (service *Service) CommitTx(ctx context.Context, transaction Transaction) error {
	currency, err := payments.PaymentCurrencyFromString(transaction.Currency)
	if err != nil {
		return ValidationError.Wrap(err)
	}

	transactions, err := service.payments.GetByCurrency(currency)
	if err != nil {
		return ValidationError.Wrap(err)
	}

	tx, err := transactions.Commit(ctx, payments.Transaction{
		Currency: currency,
		Amount:   transaction.Amount,
		To:       transaction.To,
	})
	if err != nil {
		if payments.ValidationError.Has(err) {
			return ValidationError.Wrap(err)
		}
		return Error.Wrap(err)
	}

	err = service.txDB.Commit(ctx, tx)

	return Error.Wrap(err)
}
