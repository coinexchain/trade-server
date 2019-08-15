# trade-server websocket 订阅

## General WSS information

* websocket的链接端口：wss://stream.coinexchain.com:8080
* 订阅单个信息： /ws/<streamName>
* 订阅组合信息
* 所有的订阅信息符号必须是小写; 
* 每个websocket链接的有效时间 : 

Note : 此处的实现方案还在确定中，针对链接的建立可能会有修改；

## Detailed Stream information

### 区块的确认信息

每次链上确认一个区块时，会将该区块的高度、时间戳、哈希信息推出.

**Stream Name** : blockinfo

**Payload** : 

```json
"event": "blockinfo",		// event type

{	
	"height": 162537, 			// height
	"time": 673571293,			// unix time second
	"hash": "000000000000000000ac6c4c9a6c2e406ac32b53af5910039be27f669d767356" // block hash
}
```

### 被确认的交易信息

每个区块中被确认的交易信息

**Stream Name** : txinfo

**Payload** : 

```json
"e": "blockinfo",			// event type

{
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
	],								// signers 
	"serial_number": 100, 						// tx global serial number
	"msg_types": [
		"asset/MsgIssueToken"
	],								// tx messages type
	"tx_json": "dgygygyw81728673...", // raw tx json byte
	"height": 16728						// block height 
}

```

### 验证者的Slash信息

每个区块中验证者被slash的信息

**Stream Name** : slash

**Payload** : 

```json
"e": "slash",			// event type

{
	"validator": "coinex17qtadt7356l0sf0hq5fjycnflq9lnx9c6cx5k7",	// validator address
	"power":	"67.2",				// vote power
	"reason": "Double check", 		// reason
	"jailed": true					// jailed
}	
```

### 交易对的Ticker信息

交易对的ticker信息

**Stream Name** : `<symbol>@ticker`

**Payload** : 

```json
[
	{
		"market": "bch/cet",				// market
		"new": "0.986",		// new price
		"old": "1.12" 		// old price
	},
	{
		"market": "eth/cet",				// market
		"new": "0.986",		// new price
		"old": "1.12" 		// old pric	
	}
]
```

### 交易对某个精度的K线信息

获取交易对的指定精度的K线信息

k线精度 : 当前支持 minute --> 0x10, hour --> 0x20, day --> 0x30

**Stream Name**: <symbol>@kline_<internal>

**Payload**:

```json

{
	"open": "0.989",				// open price
	"close": "0.97", 				// close price
	"high": "1.29",					// high price
	"low": "0.93", 					// low price
	"total": "8652",				// total deal stock
	"unix_time":	"786367672",		// ending unix time
	"time_span": 0x10,				// Kline internal
	"market": "etc/cet"			// trading-pai
}	 	
```

### 交易对的深度信息

获取交易对的深度信息

**Stream Name**: `<symbol>@depth_<level>`

**Payload**:

```json
{
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
```

### 交易对的成交信息

订阅指定交易对 有成交的区块信息

**Stream Name**: `<symbol>@deal`

**Payload**:

```json
{
	"stock":"89678.92",			// total deal stock in block
	"money":"7736.2",				// total deal money in block
	"height":8783					// block heigh
}
```

### 交易对的订单信息

获取指定用户的相关交易对的订单信息

**Stream Name**: `<symbol>@order_<address>`

**Payload**:

```json
// create order info
{
	"order_id": "coinex1ughhs0eyames355v4tzq5nx2g806p55rna0d2x-9",		// order id
	"sender": "coinex1ughhs0eyames355v4tzq5nx2g806p55rna0d2x", 		// order sender
	"trading_pair":	"eth/cet",				// trading-pair
	"order_type": 2,			// order type; Limit Order
	"price": "0.73", 			// order price
	"quantity": 8763892, 	// order quantity
	"side":	1, 					// order side; BUY / SELL
	"time_in_force": 3,		// GTC / IOC	"feature_fee": 21562，	// order feature fee; sato.CET as the unit
	"height": 2773,				// block height
	"frozen_fee": 782553,	// order frozen fee; sato.CET as the unit
	"freeze": 836382			// freeze sato.CET amount
}

// fill order info
{
	"order_id": "coinex1ughhs0eyames355v4tzq5nx2g806p55rna0d2x-9",		// order id
	"trading_pair":	"eth/cet",				// trading-pair
	"height": 2773,				// block height
	"side":	1, 					// order side; BUY / SELL
	"price": "0.73", 			// order price
	"freeze": 836382,			// freeze sato.CET amount
	"deal_stock": 773,			// order deal stock
	"deal_money": 726,			// order deal money
	"curr_stock": 8262,		// order remain stock
	"curr_money": 7753		// order remain money
}

// cancel order info 
{
	"order_id": "coinex1ughhs0eyames355v4tzq5nx2g806p55rna0d2x-9",		// order id
	"trading_pair":	"eth/cet",				// trading-pair
	"height": 2773,				// block height
	"side":	1, 					// order side; BUY / SELL
	"price": "0.73", 			// order price
	"del_reason": "munual delete order", 
	"used_commission": 37253, 	// order used commission
	"left_stock": 7753, 				// order left stock
	"remain_amount": 7762,			// order remain amount
	"deal_stock": 773,			// order deal stock
	"deal_money": 722		// order deal money
}
```

### 股吧评论信息

获取指定token的股吧信息；

**Stream Name**: `<tokenSymbol>@comment`

**Payload**:

```json
{
	"id": 2,
	"height": 2773,				// block height
	"sender": "coinex1ughhs0eyames355v4tzq5nx2g806p55rna0d2x", 		// comment sender
	"token": "eth",				// token symbol
	"donation": 112,			// donation amount
	"title": "The bese value token", 	// comment title
	"content": "Its price will grow", 		// "content"
	"content_type": 2, 			// content type
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

```

Content Type | value |
-----------|------------|
RawBytes | 6
UTF8Text | 3
HTTP | 2 
Magnet | 1
IPFS | 0


### bancor合约的信息

获取bancor合约信息

**Stream Name**: `<symbol>@bancor-info`

**Payload**:

```json
// bancor info

{
	"sender": "coinex1ughhs0eyames355v4tzq5nx2g806p55rna0d2x",
	"stock": "set",
	"money": "cet",
	"init_price": "67351", 				// bancor init price
	"max_supply": "27367129", 			// bancor max supply 
	"max_price":	"3.212", 			// bancor max price
	"price": "1.12", 					// bancor curr price
	"stock_in_pool": "322",			// stock in pool
	"money_in_pool": "21636", 		// money in pool
	"enable_cancel_time": 76637128,			// bancor enable cancel time 
	"block_height": 7632 			// block height 
}

```

### 订阅bancor 合约的成交信息

获取bancor合约的成交信息

**Stream Name**: `<symbol>@bancor-trade`

**Payload**:

```json
// bancor trade
{
	"sender": "coinex1ughhs0eyames355v4tzq5nx2g806p55rna0d2x",
	"stock": "set",
	"money": "cet",
	"amount": 27372,
	"side": "1",
	"monet_limit": 231,
	"transaction_price": "2123.2", 
	"block_height": 3631
}
```

### 账户的金额变动信息

获取用户指定unix时间戳前，所有的收入记录； 

**Stream Name**: `<address>@income_<unix_time>`

**Payload**:

```json
[
	{
		"type": "bankx/MsgSend",			// msg type
		"amount": "21672.21" 		//	amount	
	},
	{
		"type": "bankx/MsgSend",			// msg type
		"amount": "21672.21" 		//	amount	
	}
]
```


### 用户的Redelegation 信息

获取用户 Redelegation 的信息

**Stream Name**: `<address>@redelegation`

**Payload**:

```json
// begin redelegation

{
	"height": 12322,
	"delegator": "coinex1ughhs0eyames355v4tzq5nx2g806p55rna0d2x",
	"src": "coinex1ughhs0eyames355v4tzq5nx2g806p55rna0d2x",
	"dst": "coinex1ughhs0eyames355v4tzq5nx2g806p55rna0d2x",
	"amount": "32313",
	"completion_time": "132432"
}

// complete redelegation

{
	"height": 12322,
	"delegator": "coinex1ughhs0eyames355v4tzq5nx2g806p55rna0d2x",
	"src": "coinex1ughhs0eyames355v4tzq5nx2g806p55rna0d2x",
	"dst": "coinex1ughhs0eyames355v4tzq5nx2g806p55rna0d2x",
}
```


### 用户的unbonding信息

获取用户的unbonding 信息

**Stream Name**: `<address>@unbonding`

**Payload**:

```json
// begin unbonding

{
	"height": 12322,
	"delegator": "coinex1ughhs0eyames355v4tzq5nx2g806p55rna0d2x",
	"validator": "coinex1ughhs0eyames355v4tzq5nx2g806p55r323x",
	"amount": "9023",
	"completion_time": "1234323",
}

// complete unbonding 

{
	"height": 22384,
	"delegator": "coinex1ughhs0eyames355v4tzq5nx2g806p55rna0d2x",
	"validator": "coinex1ughhs0eyames355v4tzq5nx2g806p55rna123x",
}
```

### 用户的unblock 信息

获取用户的延迟转账到期信息

**Stream Name**: `<address>@unlock`

**Payload**:

```json
{
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
```		



