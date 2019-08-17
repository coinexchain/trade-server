# trader-server
DEX trade backend

[![Build Status](https://travis-ci.com/coinexchain/trade-server.svg?token=SzpkQ9pqByb4D3AFKW7z&branch=master)](https://travis-ci.com/coinexchain/trade-server)
[![Coverage Status](https://coveralls.io/repos/github/coinexchain/trade-server/badge.svg?t=OJj2bl)](https://coveralls.io/github/coinexchain/trade-server)

## Quick start
Copy from default configuration file
```bash
cp config.toml.default config.toml
```
Modify the configuration, and start the server
```bash
nohup trade-server -c config.toml &
```
