This is a web service that allow us to transfer eth and btc assets.

It has only 1 web api handler to do that.

This service also contains some CLI that will allow usto generate config, db schema and run service.

## Project structure

`cmd` - contains CLI application.

`internal` - contains some internal programming modules.

`console` - contains web server and business logic.

`paxfuldb` - database implementation.

`payments` - contains different crypto currency implementations with all related logic.

`peer.go` - actually, our program wrapper.

### cmd package

Our CLI has only 2 commands - `setup` and `run`.

`setup` command should be used to create config file, database and schema - `paxful setup`.

`run` command will run web server - `paxful run`.

### internal package

This package contains the only programming module - logger.
Also, all "helpers" and "utils" functions should be placed here.

### console package

paxful web server has 1 endpoint with appropriate handler:

```
router.Handle("/", http.HandlerFunc(server.CommitTx)).Methods(http.MethodPost)
```

`CommitTx` - is a web api handler that is used to commit a transaction.

### Configuration

Here is all possible configurations for paxful payment service:

```
// Config is the global configuration to interact with paxful payment service through CLI.
Config{
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
```
