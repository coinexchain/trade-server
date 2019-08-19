﻿# websocket 订阅的信息

## 订阅income信息
`income:<address>`

订阅示例：`income:coinex1q7nm8r8v2wzhrq8wxel9qpc832vnhmf5kvxuxs`

应答:
```
2019/08/19 17:54:43 recv: {"type":"income", "payload":"{"signers":["coinex1q7nm8r8v2wzhrq8wxel9qpc832vnhmf5kvxuxs"],"transfers":[{"sender":"coinex1q7nm8r8v2wzhrq8wxel9qpc832vnhmf5kvxuxs","recipient":"coinex16ur229a4xkj9e0xu06nqge9c23y70g7sl5vj98","amount":"600000000cet"}],"serial_number":24638,"msg_types":["MsgSend"],"tx_json":"{\"msg\":[{\"from_address\":\"coinex1q7nm8r8v2wzhrq8wxel9qpc832vnhmf5kvxuxs\",\"to_address\":\"coinex16ur229a4xkj9e0xu06nqge9c23y70g7sl5vj98\",\"amount\":[{\"denom\":\"cet\",\"amount\":\"600000000\"}],\"unlock_time\":0}],\"fee\":{\"amount\":[{\"denom\":\"cet\",\"amount\":\"200000000\"}],\"gas\":1000000},\"signatures\":[{\"pub_key\":[2,248,63,173,181,253,229,228,228,124,52,132,112,65,124,37,0,19,131,25,25,77,63,79,7,210,0,25,44,126,174,88,242],\"signature\":\"skkAap2ltmG2VNulgIuTLPnfuh0wxy8Id4iwzc/i+QFS56pCY28h9cIGxz9ajQxGupm6C4NWLq87wU4l4sQbEQ==\"}],\"memo\":\"send money from one account to another\"}","height":2367}"}
```

## 订阅txs信息

`txs:<address>`

订阅示例： `txs:coinex1q7nm8r8v2wzhrq8wxel9qpc832vnhmf5kvxuxs`

应答：

```
2019/08/19 19:53:00 recv: {"type":"txs", "payload":"{"signers":["coinex16ur229a4xkj9e0xu06nqge9c23y70g7sl5vj98"],"transfers":[{"sender":"coinex16ur229a4xkj9e0xu06nqge9c23y70g7sl5vj98","recipient":"coinex1x6u3jacjqaqdfwvj4fs7adtumpsv2y6p6258jl","amount":"600000000cet"}],"serial_number":49984,"msg_types":["MsgSend"],"tx_json":"{\"msg\":[{\"from_address\":\"coinex16ur229a4xkj9e0xu06nqge9c23y70g7sl5vj98\",\"to_address\":\"coinex1x6u3jacjqaqdfwvj4fs7adtumpsv2y6p6258jl\",\"amount\":[{\"denom\":\"cet\",\"amount\":\"600000000\"}],\"unlock_time\":0}],\"fee\":{\"amount\":[{\"denom\":\"cet\",\"amount\":\"200000000\"}],\"gas\":1000000},\"signatures\":[{\"pub_key\":[3,147,127,6,163,84,111,186,5,197,115,8,105,97,215,142,146,25,9,250,99,108,15,250,104,136,115,35,8,148,156,158,41],\"signature\":\"F3Zo4xzmc8skzpldzEmcUkpgwnQ65cnoUv7SQh5bLNMd9uNDr08/YRwFzJqHExtiBAZG9VJMHkRnVtUDVPq2RQ==\"}],\"memo\":\"send money from one account to another\"}","height":3977}"}
```

## 订阅blockinfo 信息

`blockinfo`

示例：`blockinfo`

应答：

```
2019/08/19 19:48:57 recv: {"type":"blockinfo", "payload":"{"height":3873,"timestamp":"2019-08-19T11:48:54.507868Z","last_block_hash":"0CB05CB285A61837248F3133A3EBD7C32AD61A5B3729B616274C506E1A886EED"}"}
```

## 订阅Ticker 信息

`ticker:<trading-pair>`

示例： `ticker:eth1/cet`

应答：

```
2019/08/19 21:00:10 recv: {"type":"ticker","payload":{"tickers":[{"market":"sdu1/cet","new":"4.799999999600122541","old":"4.799999999600122541"}]}}
```

示例：`order:coinex16ur229a4xkj9e0xu06nqge9c23y70g7sl5vj98`

## 订阅 交易对的成交(deal)信息

`deal:<trading-pair>`

示例：`deal:sdu1/cet`

应答：

```
2019/08/19 20:32:04 recv: {"type":"deal", "payload":"{"order_id":"coinex1c87uzudwwrgjmq5zude7k0s5t7r8cl6333v6uv-12524","trading_pair":"sdu1/cet","height":4421,"side":2,"price":"4.300000000000000000","left_stock":0,"freeze":0,"deal_stock":587912094,"deal_money":2939560470,"curr_stock":587912094,"curr_money":2939560470}"}
```

## 订阅order信息

`order:<address>`

示例：`order:coinex16ur229a4xkj9e0xu06nqge9c23y70g7sl5vj98`

应答：

```
// 订单创建
2019/08/19 20:32:06 recv: {"type":"order", "payload":"{"order_id":"coinex16ur229a4xkj9e0xu06nqge9c23y70g7sl5vj98-11678","sender":"coinex16ur229a4xkj9e0xu06nqge9c23y70g7sl5vj98","trading_pair":"sdu1/cet","order_type":2,"price":"3.100000000000000000","quantity":563125147,"side":2,"time_in_force":3,"feature_fee":0,"height":4422,"frozen_fee":2815626,"freeze":563125147}"}

// 订单成交
2019/08/19 20:32:04 recv: {"type":"order", "payload":"{"order_id":"coinex16ur229a4xkj9e0xu06nqge9c23y70g7sl5vj98-11667","trading_pair":"sdu1/cet","height":4421,"side":1,"price":"5.300000000000000000","left_stock":0,"freeze":174882868,"deal_stock":582942891,"deal_money":2914714455,"curr_stock":582942891,"curr_money":2914714455}"}

// 订单删除

2019/08/19 20:32:06 recv: {"type":"order", "payload":"{"order_id":"coinex16ur229a4xkj9e0xu06nqge9c23y70g7sl5vj98-11679","trading_pair":"sdu1/cet","height":4422,"side":1,"price":"7.000000000000000000","del_reason":"The order was fully filled","used_commission":2598017,"left_stock":0,"remain_amount":935285865,"deal_stock":519603258,"deal_money":2701936941}"}

```

## 订阅订单深度信息

`depth:<trading-pair>`

示例：`depth:sdu1/cet`

应答：

```
2019/08/19 20:32:05 recv: {"type":"depth","payload":{"trading_pair":"sdu1/cet","bids":null,"asks":[{"p":"5.100000000000000000","a":"-4400671459"},{"p":"3.100000000000000000","a":"0"},{"p":"9.500000000000000000","a":"1419700229"},{"p":"6.700000000000000000","a":"-115123610"},{"p":"8.200000000000000000","a":"-2236331597"}}}

2019/08/19 20:32:05 recv: {"type":"depth","payload":{"trading_pair":"sdu1/cet","bids":[{"p":"0.100000000000000000","a":"2771772552"},{"p":"7.300000000000000000","a":"0"},{"p":"3.100000000000000000","a":"-3323345383"},{"p":"3.500000000000000000","a":"-2417677642"},{"p":"1.500000000000000000","a":"-1839243231"},{"p":"0.800000000000000000","a":"-553828543"},{"p":"7.000000000000000000","a":"0"}}}
```

## 订阅K线信息

`kline:<trading-pair>:<level>`

示例：`kline:sdu1/cet:16`

应答：
```
2019/08/19 21:09:10 recv: {"type":"kline", "payload":"{"open":"5.199999998894717703","close":"4.899999999013645027","high":"5.300000000000000000","low":"4.899999991850623725","total":"507280335742","unix_time":1566220136,"time_span":16,"market":"sdu1/cet"}"}
2019/08/19 21:09:11 recv: {"type":"ticker","payload":{"tickers":[{"market":"sdu1/cet","new":"4.899999999013645027","old":"4.899999999013645027"}]}}
```