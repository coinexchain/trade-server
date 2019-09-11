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




---------------------------

## Test Order Depth Update

1. 运行trade-server
 
    * Run : `./trade-server`
2. 运行websocket测试程序，接收websocket的推送 
    
    * `go run websocket_examples.go`   
    
    ```
    2019/09/11 16:29:00 recv: {"type":"depth", "payload":{"trading_pair":"abcd/cet","bids":[{"p":"0.005200000000000000","a":"0"}],"asks":null}}
    2019/09/11 16:29:00 recv: {"type":"depth", "payload":{"trading_pair":"abcd/cet","bids":[{"p":"0.005200000000000000","a":"0"}],"asks":null}}
    2019/09/11 16:29:00 recv: {"type":"depth", "payload":{"trading_pair":"abcd/cet","bids":[{"p":"0.005200000000000000","a":"0"}],"asks":null}}
    ```
3. 向kafka中输入测试数据

    * `kafka-console-producer --broker-list localhost:9092 --topic coinex-dex  --property parse.key=true --property key.separator=# < data.txt`
    * 该文本文件包含3笔IOC的买单；将这些数据喂给kafka；然后由trade-server进行消费.

4. 测试Rest接口的查询

    * `curl -X GET "http://localhost:8000/market/depths?market=abcd%2Fcet&count=10" -H "accept: application/json"`
    *  Return value : `{"sell":[],"buy":[]}`