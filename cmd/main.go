// Copyright (C) 2020 Creditor Corp. Group.
// See LICENSE for copying information.

package main

import (
	"context"
	"database/sql"
	"os"
	"paxful/console/server"
	"paxful/payments/paymentsbtc"
	"paxful/payments/paymentsconfig"
	"paxful/payments/paymentseth"

	"github.com/spf13/cobra"
	"github.com/zeebo/errs"

	"paxful"
	"paxful/internal/logger/zaplog"
	"paxful/paxfuldb"
)

var Error = errs.Class("paxful payments CLI error")

// Config is the global configuration to interact with paxful payment service through CLI.
type Config struct {
	DatabaseURL string
	paxful.Config
}

// commands
var (
	// payments root cmd.
	rootCmd = &cobra.Command{
		Use:   "payments",
		Short: "CLI for interacting with paxful payment service",
	}

	// payments setup cmd.
	setupCmd = &cobra.Command{
		Use:         "setup",
		Short:       "setups the program config, creates database",
		RunE:        cmdSetup,
		Annotations: map[string]string{"type": "setup"},
	}
	runCmd = &cobra.Command{
		Use:         "run",
		Short:       "runs the program",
		RunE:        cmdRun,
		Annotations: map[string]string{"type": "run"},
	}
	runCfg   Config
	setupCfg Config
)

func init() {
	runCfg = getRunConfig()
	setupCfg = getSetupConfig()
	rootCmd.AddCommand(runCmd)
	rootCmd.AddCommand(setupCmd)
}

func main() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func cmdRun(cmd *cobra.Command, args []string) (err error) {
	ctx := context.Background()
	log := zaplog.NewLog()

	db, err := paxfuldb.NewDatabase(runCfg.DatabaseURL)
	if err != nil {
		log.Error("Error starting master database on paxful payment service", Error.Wrap(err))
		return Error.Wrap(err)
	}
	defer func() {
		err = errs.Combine(err, db.Close())
	}()

	peer, err := paxful.NewPeer(log, db, runCfg.Config)
	if err != nil {
		log.Error("Error starting paxful payment service", Error.Wrap(err))
		return Error.Wrap(err)
	}

	runError := peer.Run(ctx)
	closeError := peer.Close()
	return Error.Wrap(errs.Combine(runError, closeError))
}

func cmdSetup(cmd *cobra.Command, args []string) (err error) {
	// TODO: should also create config with default values.
	ctx := context.Background()
	log := zaplog.NewLog()

	conn, err := sql.Open("postgres", setupCfg.DatabaseURL)
	if err != nil {
		log.Error("could not connect to database server", Error.Wrap(err))
		return Error.Wrap(err)
	}
	defer func() {
		err = errs.Combine(err, conn.Close())
	}()

	createDBQuery := "CREATE DATABASE paxfuldb;"

	_, err = conn.ExecContext(ctx, createDBQuery)
	if err != nil {
		log.Error("can not create paxfuldb", Error.Wrap(err))
		return Error.Wrap(err)
	}

	createTableQuery :=
		`
		CREATE TABLE transactions (
			id            TEXT   NOT NULL,
			currency 	  TEXT   NOT NULL,
			amount        bigint NOT NULL,
			fee           bigint NOT NULL,
			fromAddress   TEXT   NOT NULL,
			toAddress     TEXT   NOT NULL,
			created_at    timestamp with time zone NOT NULL
		);
		`

	_, err = conn.ExecContext(ctx, createTableQuery)
	if err != nil {
		log.Error("can not create transactions table", Error.Wrap(err))
		return Error.Wrap(err)
	}

	return err
}

// getRunConfig reads config from specified source
// TODO: should read from file, but i don't have enough time.
func getRunConfig() Config {
	return Config{
		DatabaseURL: `
			host = localhost
			dbname = paxfuldb
			port = 5432
			user = postgres
			password = 123456
			connect_timeout = 2
			sslmode = disable
		`,
		Config: paxful.Config{
			Server: server.Config{
				Address: ":8081",
			},
			Payments: paymentsconfig.Config{
				CommissionPercent: 1.5,
				Ethereum: paymentseth.Config{
					URL:           "https://rinkeby.infura.io/v3/b6772462cc364bedbfaecd8acaf6982b",
					PrivateKey:    "178d6c54654274d86948b1ac351eae25d8b0a19f6f9508f45ec138ed2894f13e",
					GasLimit:      21000,
					GasPriceInWei: 30000000000,
				},
				Bitcoin: paymentsbtc.Config{
					URL:           "",
					PrivateKey:    "",
					GasLimit:      0,
					GasPriceInWei: 0,
				},
			},
		},
	}
}

// getRunConfig reads config from specified source
// TODO: should read from file, but i don't have enough time.
func getSetupConfig() Config {
	return Config{
		DatabaseURL: `
			host = localhost
			port = 5432
			user = postgres
			password = 123456
			connect_timeout = 2
			sslmode = disable
		`,
		Config: paxful.Config{
			Server: server.Config{},
			Payments: paymentsconfig.Config{
				Ethereum: paymentseth.Config{
					URL:           "",
					PrivateKey:    "",
					GasLimit:      0,
					GasPriceInWei: 0,
				},
				Bitcoin: paymentsbtc.Config{
					URL:           "",
					PrivateKey:    "",
					GasLimit:      0,
					GasPriceInWei: 0,
				},
			},
		},
	}
}
