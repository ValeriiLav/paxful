// Copyright (C) 2020 Creditor Corp. Group.
// See LICENSE for copying information.

package paymentsconfig

import (
	"paxful/payments/paymentsbtc"
	"paxful/payments/paymentseth"
)

// Config defines global payments config.
type Config struct {
	CommissionPercent float64
	Ethereum          paymentseth.Config
	Bitcoin           paymentsbtc.Config
}
