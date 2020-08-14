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
{
    "databaseUrl": "your database url",
    "config": {
        "server": {
            "address": ":8081"
        },
        "payments": {
            "commissionPercent": 1.5,
            "ethereum": {
                "url": "https://rinkeby.infura.io/v3/{projectID}",
                "privateKey": "ethereum-private-key",
                "gasLimit": 21000,
                "gasPriceInWei": 30000000000
            },
            "bitcoin": {
                "url": "qweqw",
                "privateKey": "qweqwe",
                "gasLimit": 1,
                "gasPriceInWei": 2
            }
        }
    }
}
```

## How to run
unfortunatelly, this solution don't have any containerization 

so, golang 1.13 should be installed
psql server should be run

to change config values you are able to modify
paxful/cmd/main.go 
line 140, getRunConfig function

curl request to test:
curl --location --request POST 'localhost:8081' --header 'Content-Type: application/json' --data '{"currency": "eth", "amount": 10, "to":"0x89205A3A3b2A69De6Dbf7f01ED13B2108B2c43e7"}'

