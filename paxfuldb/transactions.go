// Copyright (C) 2020 Creditor Corp. Group.
// See LICENSE for copying information.

package paxfuldb

import (
	"context"
	"database/sql"
	"github.com/zeebo/errs"

	"paxful/payments"
)

// ensures that transactions implements payments.Transactions.
var _ payments.TransactionsDB = (*transactions)(nil)

// TransactionDBError in the error class that indicates about TransactionDB error.
var TransactionDBError = errs.Class("TransactionDB error")

// transactions is a postgres implementations of a payments.TransactionsDB.
//
// architecture: Database
type transactions struct {
	db *sql.DB
}

// Commit is used to create new transaction record in TransactionDB.
func (transactions *transactions) Commit(ctx context.Context, transaction payments.Transaction) error {
	statement := `INSERT INTO transactions (id, currency, amount, fee, fromAddress, toAddress, created_at) VALUES ($1, $2, $3, $4, $5, $6, $7);`

	_, err := transactions.db.ExecContext(ctx, statement, transaction.ID, transaction.Currency, transaction.Amount, transaction.Fee, transaction.From, transaction.To, transaction.CreatedAt)

	return TransactionDBError.Wrap(err)
}


// List is used to return all transactions.
func (transactions *transactions) List(ctx context.Context) ([]payments.Transaction, error) {
	statement := `SELECT * FROM transactions;`

	var transactionList []payments.Transaction

	rows, err := transactions.db.QueryContext(ctx, statement)
	if err != nil {
		return nil, TransactionDBError.Wrap(err)
	}
	defer func() { err = errs.Combine(err, rows.Close()) }()

	for rows.Next() {
		transaction := payments.Transaction{}

		err := rows.Scan(&transaction.ID, &transaction.Currency, &transaction.Amount, &transaction.Fee, &transaction.From, &transaction.To, &transaction.CreatedAt);
		if err != nil {
			return nil, TransactionDBError.Wrap(err)
		}

		transactionList = append(transactionList, transaction)
	}
	if err = rows.Err(); err != nil {
		return nil, TransactionDBError.Wrap(err)
	}

	return transactionList, nil
}

