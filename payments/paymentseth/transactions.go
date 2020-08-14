// Copyright (C) 2020 Creditor Corp. Group.
// See LICENSE for copying information.

package paymentseth

import (
	"context"
	"crypto/ecdsa"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/zeebo/errs"

	"paxful/internal/logger"
	"paxful/payments"
)

// ensures that transactions implements payments.Transactions.
var _ payments.Transactions = (*transactions)(nil)

// Error is an error class for internal ethereum transaction service error.
var Error = errs.Class("ethereum transaction error")

// Config stores needed information for eth payment service initialization.
type Config struct {
	URL           string // "https://mainnet.infura.io"
	PrivateKey    string
	GasLimit      uint64
	GasPriceInWei int64
}

// transactions is an ETH implementation of paxful payment service.
type transactions struct {
	log    logger.Logger
	config Config

	commissionPercent float64

	eth *ethclient.Client
}

// NewClient is a constructor for a ETH client.
func NewTransactions(log logger.Logger, config Config, commissionPercent float64) (payments.Transactions, error) {
	client, err := ethclient.Dial(config.URL)
	if err != nil {
		return nil, Error.Wrap(err)
	}

	return &transactions{
		log:               log,
		eth:               client,
		config:            config,
		commissionPercent: commissionPercent,
	}, nil
}

// Commit injects a signed transaction into the pending pool for execution.
func (t *transactions) Commit(ctx context.Context, tx payments.Transaction) (payments.Transaction, error) {
	privateKey, err := crypto.HexToECDSA(t.config.PrivateKey)
	if err != nil {
		return payments.Transaction{}, Error.Wrap(err)
	}

	publicKeyECDSA, ok := privateKey.Public().(*ecdsa.PublicKey)
	if !ok {
		return payments.Transaction{}, payments.ValidationError.New("public key could not be converted to ECDSA")
	}

	from := crypto.PubkeyToAddress(*publicKeyECDSA)
	nonce, err := t.eth.PendingNonceAt(ctx, from)
	if err != nil {
		return payments.Transaction{}, Error.Wrap(err)
	}

	// we assume that transaction amount field were in "wei" currency.
	weiAmount := FloatToBigInt(t.applyCommission(tx.Amount))

	// in case when we don't want to configure gas price manually.
	gasPrice := big.NewInt(t.config.GasPriceInWei)
	if t.config.GasPriceInWei == 0 {
		gasPrice, err = t.eth.SuggestGasPrice(ctx)
		if err != nil {
			return payments.Transaction{}, Error.Wrap(err)
		}
	}

	if !common.IsHexAddress(tx.To) {
		return payments.Transaction{}, payments.ValidationError.New("receiver address is not valid Hex address")
	}

	unsignedTx := types.NewTransaction(nonce, common.HexToAddress(tx.To), weiAmount, t.config.GasLimit, gasPrice, nil)

	// signing transaction.
	chainID, err := t.eth.NetworkID(ctx)
	if err != nil {
		return payments.Transaction{}, Error.Wrap(err)
	}

	signedTx, err := types.SignTx(unsignedTx, types.NewEIP155Signer(chainID), privateKey)
	if err != nil {
		return payments.Transaction{}, Error.Wrap(err)
	}

	// transferring assets.
	err = t.eth.SendTransaction(ctx, signedTx)
	if err != nil {
		return payments.Transaction{}, Error.Wrap(err)
	}

	tx.ID = signedTx.Hash().String()
	tx.CreatedAt = time.Now().UTC()
	tx.From = from.String()
	tx.Fee = gasPrice.Int64()

	return tx, nil
}

// applyCommission calculates new amount after applying commission.
func (t *transactions) applyCommission(amount float64) float64 {
	return amount / 100 * t.commissionPercent
}

// should be placed to internal pkg.
func FloatToBigInt(val float64) *big.Int {
	bigval := new(big.Float)
	bigval.SetFloat64(val)

	coin := new(big.Float)
	coin.SetInt(big.NewInt(1000000000000000000))

	bigval.Mul(bigval, coin)

	result := new(big.Int)
	bigval.Int(result)

	return result
}
