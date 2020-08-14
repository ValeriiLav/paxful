// Copyright (C) 2020 Creditor Corp. Group.
// See LICENSE for copying information.

package paxfuldb

import (
	"database/sql"

	_ "github.com/lib/pq"
	"github.com/zeebo/errs"

	"paxful"
	"paxful/payments"
)

// ensures that database implements paxful.DB.
var _ paxful.DB = (*database)(nil)

var (
	// Error is the default paxfuldb error class.
	Error = errs.Class("paxfuldb error")
)

// database combines access to different database tables with a record
// of the db driver, db implementation, and db source URL.
//
// architecture: Master Database
type database struct {
	db *sql.DB
}

// NewDatabase returns paxful.DB postgresql implementation.
func NewDatabase(databaseURL string) (paxful.DB, error) {
	conn, err := sql.Open("postgres", databaseURL)
	if err != nil {
		return nil, Error.Wrap(err)
	}

	return &database{db: conn}, nil
}

// Transactions provides access to Transactions store.
func (db *database) Transactions() payments.TransactionsDB {
	return &transactions{
		db: db.db,
	}
}

// Close closes underlying db connection.
func (db *database) Close() error {
	return Error.Wrap(db.db.Close())
}
