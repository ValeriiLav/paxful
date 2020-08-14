// Copyright (C) 2020 Creditor Corp. Group.
// See LICENSE for copying information.

package console

// Transaction hold information needed to create transaction.
type Transaction struct {
	Currency string  `json:"currency"`
	Amount   float64 `json:"amount"`
	To       string  `json:"to"`
}
