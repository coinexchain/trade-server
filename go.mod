module github.com/coinexchain/trade-server

go 1.12

require (
	github.com/coinexchain/dex v0.0.11
	github.com/cosmos/cosmos-sdk v0.36.0-rc4
	github.com/emirpasic/gods v1.12.0
	github.com/gorilla/mux v1.7.0
	github.com/pelletier/go-toml v1.2.0
	github.com/tendermint/tm-db v0.1.1
)

#replace github.com/coinexchain/dex v0.0.11 => ../dex
