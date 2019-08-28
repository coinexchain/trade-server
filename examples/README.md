## install kafka on mac

`brew install kafka`

## run zookeeper

`zookeeper-server-start /usr/local/etc/kafka/zookeeper.properties &`

## run kafka

`kafka-server-start /usr/local/etc/kafka/server.properties &`

## compile and run trade-server

`cd ..`

Compile : `github.com/coinexchain/trade-server/`

Run : `./trade-server`

## run examples to recv data from websocket connection

`cd examples`

`go run websocket_examples.go` 

## Feed data to the trade-server 

`kafka-console-producer --broker-list localhost:9092 --topic coinex-dex  --property parse.key=true --property key.separator=# < ../docs/dex_msgs_data.txt`
