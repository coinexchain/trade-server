# trader-server
DEX trade backend

[![Build Status](https://travis-ci.com/coinexchain/trade-server.svg?token=SzpkQ9pqByb4D3AFKW7z&branch=master)](https://travis-ci.com/coinexchain/trade-server)
[![Coverage Status](https://coveralls.io/repos/github/coinexchain/trade-server/badge.svg?t=OJj2bl)](https://coveralls.io/github/coinexchain/trade-server)

## How to run trade-server
1. Copy ``config.toml.default`` to ``config.toml``.
2. Update the configuration.
3. Start the server
```bash
nohup trade-server -c config.toml &
```
