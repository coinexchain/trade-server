# trade-server websocket 订阅

## 链接

将您的 websocket 客户端连接到 `ws://localhost:8000/ws`

## 所有指令

在链接建立后，通过发送下列格式的指令数据，来订阅、解订阅相关的topic.

基本的指令发送格式

`{"op":"<command>", "args":["args1", "args2", "args3", ...], "depth": <10> }`

*	订阅
	*  subscribe
	*  unsubscribe
* 心跳
	* Ping
	
	
## 订阅

trader-server 是用来订阅实时数据的，一旦链接成功，会获取到订阅主题的最新信息推送，它是获取最新数据的最好方法。

订阅主题，采用下列这种方式；

*	链接已建立后，通过发送下列信息至`trade-server`，订阅一个新的主题，使用下列格式发送请求
	* `{"op":"subscribe", "args":[<subscriptionTopic>, ...]}`
通过发送订阅主题数组，一次可订阅多个主题。

当前所有的主题订阅，均无需进行身份验证。

当订阅的某个topic需要携带参数时，	可以使用`:`号分隔topic名称与它的参数；

*	如：`kline:etc/cet:1min` ; 该topic的意思为：订阅 etc/cet的1分钟K线数据.

> [golang 客户端示例](https://github.com/coinexchain/trade-server/blob/master/examples/websocket_examples.go)

> [响应数据示例](https://github.com/coinexchain/trade-server/blob/master/docs/websocket-data-examples.md)

## 响应格式

websocket的响应可能含有以下三种类型：

`Success`(成功订阅一个主题后的响应)
`{"subscribe": subscriptionName, "success": true}`

`Error`(格式错误的请求的响应)
`{"error": errorMessage}`

`数据响应`(当订阅的数据被推送时): 

```json
{
	"type": "blockinfo",   		// 数据应答类型
	"payload":{
        ...
        // 数据应答信息
        ...	
	}
}
```
所有的数据消息推送都有一个 `type`属性，用来标识消息类型，以便对它进行响应的处理.
	
## 主题列表

### 区块的确认信息

每次链上确认一个区块时，会将该区块的高度、时间戳、哈希信息推出.

**SubscriptionTopic** : `blockinfo`

**Response** : 

```json

{
	"type": "blockinfo",
	"payload": {	
            "height": 162537, 			// height
            "time": 673571293,			// unix time second
            "hash": "000000000000000000ac6c4c9a6c2e406ac32b53af5910039be27f669d767356" // block hash
	}
}
```

### 被确认的交易信息

获取每个区块中指定用户签名的交易。

**SubscriptionTopic** : `txs:<address>`

**Response** : 

```json

{
	"type": "txs",
	"payload": {
            "transfers": [
                {
                    "sender": "coinex1ughhs0eyames355v4tzq5nx2g806p55rna0d2x",
                    "recipient": "coinex17qtadt7356l0sf0hq5fjycnflq9lnx9c6cx5k7",
                    "amount": "8467.863"		// amount
                },
                {
                    "sender": "coinex1ughhs0eyames355v4tzq5nx2g806p55rna0d2x",
                    "recipient": "coinex17qtadt7356l0sf0hq5fjycnflq9lnx9c6cx5k7",
                    "amount": "8467.863"		// amount
                }
            ],
            "signers": [
                "coinex1ughhs0eyames355v4tzq5nx2g806p55rna0d2x",
                "coinex17qtadt7356l0sf0hq5fjycnflq9lnx9c6cx5k7"
            ],							// signers 
            "serial_number": 100, 				// tx global serial number
            "msg_types": [
                "asset/MsgIssueToken"
            ],							// tx messages type
            "tx_json": "...",                                   // raw tx json byte
            "height": 16728				        // block height
	} 
}

```

**payload** : 参见`../swagger/swagger.yaml Tx 类型定义`

### 验证者的Slash信息

获取区块中的slash信息；

**SubscriptionTopic** : `slash`

**Response** : 

```json

{
	"type": "slash",
	"payload": {
            "validator": "coinex17qtadt7356l0sf0hq5fjycnflq9lnx9c6cx5k7",	// validator address
            "power":	"67.2",				                        // vote power
            "reason": "Double Sign", 		                                // reason
            "jailed": true					                // jailed
	}
}	
```

**payload** : 参见`../swagger/swagger.yaml Slash 类型定义`

### 交易对的Ticker信息

获取交易对的ticker信息

**SubscriptionTopic** : 

*   market : `ticker:<trading-pair>`;
*   bancor : `ticker:B:<trading-pair>`

**Response** : 

```json
{
	"type": "ticker",
	"payload": {
            "tickers": [
                {
                    "market": "bch/cet",			// market
                    "new": "0.986",		                // new price
                    "old": "1.12"                               // old price
                }
            ]       
	}
}
```

**payload** : 参见`../swagger/swagger.yaml Tickers 类型定义`

### 交易对某个精度的K线信息

获取交易对的指定精度的K线信息

k线精度 : 当前支持 1min, 1hour, 1day

**SubscriptionTopic** : 

*   market : `kline:<trading-pair>:<1min>`
*   bancor : `kline:B:<trading-pair>:<1min>`

**Response**:

```json

{
	"type": "kline",
	"payload": {
            "open": "0.989",				// open price
            "close": "0.97", 				// close price
            "high": "1.29",			        // high price
            "low": "0.93", 			        // low price
            "total": "8652",				// total deal stock
            "unix_time":	"786367672",		// ending unix time
            "time_span": 16,				// Kline internal
            "market": "etc/cet"                         // trading-pai
	}
}	 	
```

**payload** : 参见`../swagger/swagger.yaml  CandleStick 类型定义`

### 交易对的深度信息

获取交易对的深度信息

**SubscriptionTopic** : `depth:<trading-pair>:<level>`

**Response**:

```json
{
    "type": "depth_change", // depth_delta, depth_full
    "payload": {
        "trading-pair": "ethsw/cet",
        "bids": [
            {
                "price": "0.936",		// price
                "amount": "100000"		// amount
            }
        ],
        "asks": [
            {
                "price": "0.936",		// price
                "amount": "100000"		// amount
            }
        ]
    }
}
```

**支持的level ** : "0.00000001", "0.0000001", "0.000001", "0.00001", "0.0001", "0.001", "0.01", "0.1", "1", "10", "100", "all"

**应答的type，含有三种类型** : depth_delta, depth_change, depth_full；详细描述如下:

*   depth_change : 使用该应答替换掉客户端中价格相同的深度数据。
*   depth_full : 使用该应答替换掉客户端所有的深度数据
*   depth_delta : 增量的深度合并数据，需要客户端依据该应答，来计算指定价格的深度数据。


**payload** : 参见`../swagger/swagger.yaml  /market/depths 请求应答`

### 交易对的成交信息

订阅指定交易对 有成交的区块信息

**SubscriptionTopic** : `deal:<trading-pair>`

**Response**:

```json
{
    "type": "deal",
    "payload": {
        "order_id": "coinex1ughhs0eyames355v4tzq5nx2g806p55rna0d2x-9",		// order id
        "trading_pair":	"eth/cet",		      // trading-pair
        "height": 2773,				      // block height
        "side":	1, 			              // order side; BUY / SELL
        "price": "0.73", 			      // order price
        "freeze": 836382,			      // freeze sato.CET amount
        "left_stock": 7753, 		              // order left stock
        "deal_stock": 773,			      // order deal stock
        "deal_money": 726,			      // order deal money
        "curr_stock": 8262,		              // curr stock
        "curr_money": 7753,			      // curr money 
        "fill_price": "0.233"              // the order deal price
    }
}

```

**payload** : 参见`../swagger/swagger.yaml  FillOrderInfo 类型定义`


### 交易对的订单信息

获取指定用户的订单信息; 

主要含有三类订单类型：创建订单，订单成交，订单取消.

**SubscriptionTopic** : `order:<address>`

**Response**:

```json
// create order info
{
	"type": "create_order",
	"payload": {
            "order_id": "coinex1ughhs0eyames355v4tzq5nx2g806p55rna0d2x-9",	    // order id
            "sender": "coinex1ughhs0eyames355v4tzq5nx2g806p55rna0d2x", 		    // order sender
            "trading_pair":	"eth/cet",           // trading-pair
            "order_type": 2,                         // order type; Limit Order
            "price": "0.73", 			     // order price
            "quantity": 8763892, 	             // order quantity
            "side":	1, 			     // order side; BUY / SELL
            "time_in_force": 3,                      // GTC / IOC
            "feature_fee": 21562,	             // order feature fee; sato.CET as the unit
            "height": 2773,		             // block height
            "frozen_fee": 782553,	             // order frozen fee; sato.CET as the unit
            "freeze": 836382			     // freeze sato.CET amount
	}
}

// **payload** : 参见`../swagger/swagger.yaml  CreateOrderInfo 类型定义`

// fill order info
{
	"type": "fill_order",
	"payload": {
            "order_id": "coinex1ughhs0eyames355v4tzq5nx2g806p55rna0d2x-9",		// order id
            "trading_pair":	"eth/cet",		// trading-pair
            "height": 2773,				// block height
            "side":	1, 				// order side; BUY / SELL
            "price": "0.73", 			        // order price
            "freeze": 836382,			        // freeze sato.CET amount
            "fill_price": "0.233",              // the order deal price
            "left_stock": 7753, 		        // order left stock
            "deal_stock": 773,			        // order deal stock
            "deal_money": 726,			        // order deal money
            "curr_stock": 8262,		                // curr stock
            "curr_money": 7753			        // curr money 
	}
}

// **payload** : 参见`../swagger/swagger.yaml  FillOrderInfo 类型定义`

// cancel order info 
{
	"type": "cancel_order",
	"payload": {
            "order_id": "coinex1ughhs0eyames355v4tzq5nx2g806p55rna0d2x-9",		// order id
            "trading_pair":	"eth/cet",			// trading-pair
            "height": 2773,				        // block height
            "side":	1, 					// order side; BUY / SELL
            "price": "0.73",                                    // order price
            "del_reason": "munual delete order", 
            "used_commission": 37253, 	                        // order used commission
            "left_stock": 7753, 				// order left stock
            "remain_amount": 7762,			        // order remain amount
            "deal_stock": 773,			                // order deal stock
            "deal_money": 722		                        // order deal money
	}
}

// **payload** : 参见`../swagger/swagger.yaml  CancelOrderInfo 类型定义`

```

### 股吧评论信息

获取指定token的股吧信息；

**SubscriptionTopic** : `comment:<symbol>`

**Response**:

```json
{
	"type": "comment",
	"payload": {
            "id": 2,
            "height": 2773,				// block height
            "sender": "coinex1ughhs0eyames355v4tzq5nx2g806p55rna0d2x", 		// comment sender
            "token": "eth",				// token symbol
            "donation": 112,			        // donation amount
            "title": "The bese value token", 	        // comment title
            "content": "Its price will grow",           // content
            "content_type": 2, 			        // content type
            "references": [
                {
                    "id": 3,
                    "reward_target": "coinex1ughhs0eyames355v4tzq5nx2g806p55rna0d2x",
                    "reward_token": "set",	
                    "reward_amount": 726712,
                    "attitudes": [
                        1, 
                        2,
                        3
                    ]		
                }
            ]
	}
}

```

**payload** : 参见`../swagger/swagger.yaml  Comment 类型定义`

Content Type | value |
-----------|------------|
RawBytes | 6
UTF8Text | 3
HTTP | 2 
Magnet | 1
IPFS | 0


### bancor合约的信息

获取bancor合约中指定交易对的信息

**SubscriptionTopic** : `bancor:<trading-pair>`

**Response**:

```json
// bancor info

{
	"type": "bancor",
	"payload": {
            "sender": "coinex1ughhs0eyames355v4tzq5nx2g806p55rna0d2x",
            "stock": "set",
            "money": "cet",
            "init_price": "67351", 				// bancor init price
            "max_supply": "27367129", 			        // bancor max supply 
            "max_price":	"3.212", 			// bancor max price
            "price": "1.12", 	                                // bancor curr price
            "stock_in_pool": "322",			        // stock in pool
            "money_in_pool": "21636", 		                // money in pool
            "enable_cancel_time": 76637128,			// bancor enable cancel time 
            "block_height": 7632 			        // block height
	} 
}

```

**payload** : 参见`../swagger/swagger.yaml  BancorInfo 类型定义`

### 订阅指定地址的 bancor 合约的成交信息

获取bancor合约的指定地址的成交信息

**SubscriptionTopic** : `bancor-trade:<address>`

**Response**:

```json
// bancor trade
{
	"type": "bancor-trade",
	"payload": {
            "sender": "coinex1ughhs0eyames355v4tzq5nx2g806p55rna0d2x",
            "stock": "set",
            "money": "cet",
            "amount": 27372,
            "side": 1,
            "monet_limit": 231,
            "transaction_price": "2123.2", 
            "block_height": 3631
	}
}
```

**payload** : 参见`../swagger/swagger.yaml  BancorTrade 类型定义`

### 订阅bancor 合约的成交信息

**SubscriptionTopic** : `bancor-deal:<trading-pair>`

**Response**:

```json
// bancor deal
{
	"type": "bancor-trade",
	"payload": {
            "sender": "coinex1ughhs0eyames355v4tzq5nx2g806p55rna0d2x",
            "stock": "set",
            "money": "cet",
            "amount": 27372,
            "side": 1,
            "monet_limit": 231,
            "transaction_price": "2123.2", 
            "block_height": 3631
	}
}
```

**payload** : 参见`../swagger/swagger.yaml  BancorTrade 类型定义`

### 账户的金额变动信息

获取指定用户 的收入信息； 

**SubscriptionTopic** : `income:<address>`

**Response**:

```json
{
    "type": "income",
        "payload": {
            "transfers": [
                {
                    "sender": "coinex1ughhs0eyames355v4tzq5nx2g806p55rna0d2x",
                    "recipient": "coinex17qtadt7356l0sf0hq5fjycnflq9lnx9c6cx5k7",
                    "amount": "8467.863"		// amount
                },
                {
                    "sender": "coinex1ughhs0eyames355v4tzq5nx2g806p55rna0d2x",
                    "recipient": "coinex17qtadt7356l0sf0hq5fjycnflq9lnx9c6cx5k7",
                    "amount": "8467.863"		// amount
                }
            ],
            "signers": [
                "coinex1ughhs0eyames355v4tzq5nx2g806p55rna0d2x",
                "coinex17qtadt7356l0sf0hq5fjycnflq9lnx9c6cx5k7"
            ],						        // signers 
            "serial_number": 100, 				// tx global serial number
            "msg_types": [
                "asset/MsgIssueToken"
            ],					                // tx messages type
            "tx_json": "...",                                   // raw tx json byte
            "height": 16728                                     // block height
        }
}
```

**payload** : 参见`../swagger/swagger.yaml Tx 类型定义`

### 用户的Redelegation 信息

获取用户 Redelegation 的信息

**SubscriptionTopic** : `redelegation:<address>`

**Response**:

```json
{
	"type": "redelegation",
	"payload": {        
            "delegator": "coinex1ughhs0eyames355v4tzq5nx2g806p55rna0d2x",
            "src": "coinex1ughhs0eyames355v4tzq5nx2g806p55rna0d2x",
            "dst": "coinex1ughhs0eyames355v4tzq5nx2g806p55rna0d2x",
            "amount": "32313",
            "completion_time": "132432"
	}
}
```

**payload** : 参见`../swagger/swagger.yaml Redelegation 类型定义`


### 用户的unbonding信息

获取用户的unbonding 信息

**SubscriptionTopic** : `unbonding:<address>`

**Response**:

```json
{
	"type": "unbonding",
	"payload": {       
            "delegator": "coinex1ughhs0eyames355v4tzq5nx2g806p55rna0d2x",
            "validator": "coinex1ughhs0eyames355v4tzq5nx2g806p55r323x",
            "amount": "9023",
            "completion_time": "1234323"
	}
}
```

**payload** : 参见`../swagger/swagger.yaml Unbonding 类型定义`


### 用户的unblock 信息

获取用户的延迟转账到期信息

**SubscriptionTopic**: `unlock:<address>`

**Response**:

```json
{
	"type": "unlock",
	"payload": {
            "address": "coinex1ughhs0eyames355v4tzq5nx2g806p55rna0d2x",
            "unlocked": [
                {
                    "amount": "2323",
                    "denom": "cet"
                },
                {
                    "amount": "3242",
                    "denom": "eth"
                }
            ],
            "locked_coins": [
                {
                    "coin": {
                        "amount": "2323",
                        "denom": "cet"
                    },
                    "unlock_time": 212323
                },
                {
                    "coin": {
                        "amount": "2323",
                        "denom": "etv"
                    },
                    "unlock_time": 223312323
                }
            ],
            "frozen_coins": [
                {
                    "amount": "2323",
                    "denom": "cet"
                },
                {
                    "amount": "3242",
                    "denom": "eth"
                }
            ],
            "coins": [
                {
                    "amount": "2323",
                    "denom": "cet"
                },
                {
                    "amount": "3242",
                    "denom": "eth"
                }
            ],
            "height": 100003
	}
}
```

**payload** : 参见`../swagger/swagger.yaml Unlock 类型定义`
	

## 心跳

客户端发送ping消息，服务器会回复pong应答。

ping消息格式：`{"op":"ping"}`
pong应答：`{"type":"pong"}`

如果你担心你的连接被默默地终止，我们推荐你采用以下流程：

*	在接收到每条消息后，设置一个 5 秒钟的定时器。
*	如果在定时器触发收到任何新消息，则重置定时器。
*	如果定时器被触发了（意味着 5 秒内没有收到新消息），发送一个 ping 数据包。

## Unsubscribe

如果在用户程序在运行一段时间后，想解除某个topic的订阅，可以使用`unsubscribe`命令。

示例： `{"op":"unsubscribe", "args":[<subscriptionTopic>, ...]}`


























