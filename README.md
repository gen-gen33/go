
# Trading App CLI

A simple CLI-based trading application using Go and CockroachDB.

## Features
- User creation and login
- Buy and sell order placement
- Order listing
- Trading function

## Requirements
- Go 1.20 or later
- CockroachDB installed and running
- Environment variables configured in `.env` file

## Installation

1. git clone <repository-url>
2. cd trading_app_cli
2. go mod tidy
3. cockroach sql --insecure --host=localhost
4. CREATE DATABASE trading; (in cockroach)
5. cockroach start-single-node --insecure --listen-addr=localhost:26257
6. go run main.go

japanese explanation is [here](https://zenn.dev/genn_tmm/articles/70a2d9025ebdb4)