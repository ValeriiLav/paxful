// Copyright (C) 2020 Creditor Corp. Group.
// See LICENSE for copying information.

package payments

import (
	"context"
	"time"

	"github.com/zeebo/errs"
)

// ValidationError indicates that transaction data is corrupted.
var ValidationError = errs.Class("transaction validation error")

// Transactions exposes functionality to work with asset transferring.
//
// architecture: Service
type Transactions interface {
	// Commit is used to send transaction to a receiver.
	Commit(ctx context.Context, tx Transaction) (Transaction, error)
}

// TransactionsDB exposes functionality to manage transactions database.
//
// architecture: Database
type TransactionsDB interface {
	// Commit is used to create new transaction record in TransactionDB.
	Commit(ctx context.Context, tx Transaction) error
	// List is used to return all transactions.
	List(ctx context.Context) ([]Transaction, error)
}

// Transaction stores information about asset transferring.
type Transaction struct {
	ID        string          `json:"id"`
	Currency  PaymentCurrency `json:"currency"`
	Amount    float64         `json:"amount"`
	Fee       int64           `json:"fee"`
	From      string          `json:"from"`
	To        string          `json:"to"`
	CreatedAt time.Time       `json:"createAt"`
}

// TransactionStatus indicates status of transaction transferring.
type TransactionStatus int

const (
	// TransactionStatusSuccess indicates that transaction was committed successfully.
	TransactionStatusSuccess TransactionStatus = 0
)
