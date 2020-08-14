// Copyright (C) 2020 Creditor Corp. Group.
// See LICENSE for copying information.

package paxful

import (
	"context"
	"errors"
	"net"
	"paxful/payments/paymentsbtc"
	"paxful/payments/paymentseth"

	"github.com/zeebo/errs"
	"golang.org/x/sync/errgroup"

	"paxful/console"
	"paxful/console/server"
	"paxful/internal/logger"
	"paxful/payments"
	"paxful/payments/paymentsconfig"
)

// DB provides access to all databases and database related functionality.
//
// architecture: Master Database
type DB interface {
	// Transactions provides access to Transactions store.
	Transactions() payments.TransactionsDB

	// Close closes underlying db connection.
	Close() error
}

// Config is the global configuration for paxful payment service.
type Config struct {
	Server   server.Config
	Payments paymentsconfig.Config
}

// Peer is the representation of a paxful payment service.
type Peer struct {
	Log      logger.Logger
	Listener net.Listener
	Service  *console.Service
	Database DB
	Endpoint *server.Server
}

// New is a constructor for paxful payment Peer.
func NewPeer(log logger.Logger, db DB, config Config) (peer *Peer, err error) {
	peer = &Peer{
		Log:      log,
		Database: db,
	}

	eth, err := paymentseth.NewTransactions(peer.Log, config.Payments.Ethereum, config.Payments.CommissionPercent)
	if err != nil {
		return nil, err
	}
	btc, err := paymentsbtc.NewTransactions(peer.Log, config.Payments.Bitcoin, config.Payments.CommissionPercent)
	if err != nil {
		return nil, err
	}
	paymentProvider := payments.NewPaymentProvider(eth, btc)
	peer.Service = console.NewService(paymentProvider, peer.Database.Transactions())

	peer.Listener, err = net.Listen("tcp", config.Server.Address)
	if err != nil {
		return nil, err
	}

	peer.Endpoint, err = server.NewServer(peer.Log, peer.Service, config.Server, peer.Listener)
	if err != nil {
		return nil, errs.Combine(err, peer.Close())
	}

	return peer, nil
}

// Run runs SNO registration service until it's either closed or it errors.
func (peer *Peer) Run(ctx context.Context) error {
	group, ctx := errgroup.WithContext(ctx)

	// start SNO registration service as a separate goroutine.
	group.Go(func() error {
		return ignoreCancel(peer.Endpoint.Run(ctx))
	})

	return group.Wait()
}

// Close closes all the resources.
func (peer *Peer) Close() error {
	errlist := errs.Group{}

	if peer.Endpoint != nil {
		errlist.Add(peer.Endpoint.Close())
	}

	if peer.Listener != nil {
		errlist.Add(peer.Listener.Close())
	}

	return errlist.Err()
}

// we ignore cancellation and stopping errors since they are expected.
func ignoreCancel(err error) error {
	if errors.Is(err, context.Canceled) {
		return nil
	}
	return err
}
