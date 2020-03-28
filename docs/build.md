## Install rocksdb

1. download: `git clone https://github.com/facebook/rocksdb.git && cd rocksdb`
2. branch: `git checkout v6.6.4`
3. build: `mkdir build && cd build && cmake .. && make -j2` 
4. install: `sudo make install`

## Build trade-server

`go build -ldflags "-X main.ReleaseVersion=<version>" -o trade-server github.com/coinexchain/trade-server`
