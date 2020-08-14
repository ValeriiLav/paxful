// Copyright (C) 2020 Creditor Corp. Group.
// See LICENSE for copying information.

package payments

import (
	"github.com/zeebo/errs"
)

var (
	// PaymentCurrencyNotSupportedError is an error class that indicates that currency in not supported.
	PaymentCurrencyNotSupportedError = errs.Class("payment currency not supported")
)

// PaymentProvider is a factory that responsible for creating different payments
// implementations depends on currency type.
//
// architecture: Service
type PaymentProvider struct {
	paymentsETH Transactions
	paymentsBTC Transactions
}

// NewPaymentProvider is a constructor for PaymentProvider.
func NewPaymentProvider(paymentsETH Transactions, paymentsBTC Transactions) PaymentProvider {
	return PaymentProvider{
		paymentsETH: paymentsETH,
		paymentsBTC: paymentsBTC,
	}
}

// GetByCurrency will return needed implementation of Transactions depends on currency type.
func (provider *PaymentProvider) GetByCurrency(currency PaymentCurrency) (Transactions, error) {
	switch currency {
	case PaymentCurrencyETH:
		return provider.paymentsETH, nil
	case PaymentCurrencyBTC:
		return provider.paymentsBTC, nil
	default:
		return nil, PaymentCurrencyNotSupportedError.New(string(currency))
	}
}

// PaymentCurrency indicates type of supported currencies to transfer.
type PaymentCurrency string

const (
	// PaymentCurrencyETH is an ethereum currency.
	PaymentCurrencyETH PaymentCurrency = "eth"
	// PaymentCurrencyBTC is an bitcoin currency.
	PaymentCurrencyBTC PaymentCurrency = "btc"
)

// PaymentCurrencyFromString creates PaymentCurrency from string.
// returns error if currency not supported.
func PaymentCurrencyFromString(currency string) (PaymentCurrency, error) {
	switch currency {
	case string(PaymentCurrencyETH):
		return PaymentCurrencyETH, nil
	case string(PaymentCurrencyBTC):
		return PaymentCurrencyBTC, nil
	default:
		return "", PaymentCurrencyNotSupportedError.New(currency)
	}
}
