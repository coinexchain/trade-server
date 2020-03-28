# trade-server 推送的数据描述

### 订阅blockinfo 信息

字段 | json类型|描述
------|------|------|
chain_id | String | 当前链ID
height | Number | 当前块高度
timestamp | Number | 当前块的时间戳
last_block_hash | String | 上个块哈希

### 订阅income信息

字段 | 类型|描述
---|---|---
signers | []String | 签名者
transfers | []Object(TransferRecord) | 见下表
serial_number | Number | 当前交易的序号
msg_types | String | 交易内的消息类型 
tx_json | String | 交易的json序列化数据
height | Number | 当前交易的高度
hash | String  | 交易哈希
extra_info | String | 交易失败时，提供的额外信息;

**TransferRecord**

字段 | 类型|描述
---|---|---
sender | String | 交易发送者
recipient | String | 交易接收者
amount |String | 金额

### 订阅指定地址的txs信息

与上述**income信息**相同

### 订阅交易对Ticker 信息

字段 | 类型|描述
---|---|---
market | String |交易对
new | String | 当前分钟的最新的价格
old | String | 当前分钟一天前的价格 
minute_in_day |Number| ticker所处在一天的第几分钟

### 订阅交易对的成交(deal)信息

字段 | 类型|描述
---|---|---
order_id| String|订单ID
trading_pair| String|交易对
height|Number|订单成交的高度
side |Number | 订单方向；BUY:1, SELL:2
price |String | 订单的挂单价 
left_stock|Number| 订单未成交的stock
freeze|Number| 订单剩余还未成交的冻结资金
deal_stock|Number| 订单成交的stock
deal_money|Number| 订单成交的money
curr_stock|Number| 订单本次成交的stock数量
curr_money|Number| 订单本次成交的money数量
fill_price|String| 订单本次的成交价

### 订阅指定地址的order信息

order包含三种类型：create_order, fill_order, cancel_order

**create_order**

字段 | 类型|描述
---|---|---
order_id | String|订单ID
sender |String|订单创建者地址
trading_pair|String|交易对
order_type |Number|2:限价单; 当前只支持限价单。
price|String|挂单价
quantity|Number|交易的stock数量
side|Number|订单方向；BUY:1, SELL:2
time_in_force|Number| 订单类型；GTE:到期单, IOC:立即成交单
height|Number|创建订单的高度
freeze|Number|订单使用的资金
frozen_feature_fee|Number|冻结订单的最大功能费；订单删除时，按存在时间返还
frozen_commission|Number|冻结订单的最大佣金；订单删除时，按成交情况返还
tx_hash|String|交易哈希

**fill_order**

与上述**订阅交易对的成交(deal)信息**相同

**cancel_order**

字段 | 类型|描述
---|---|---
order_id|String|订单ID
trading_pair|String|交易对
height|Number|删除订单时的块高度
side|Number|订单方向；BUY:1, SELL:2
price|String|订单的挂单价
del_reason|String|订单的删除原因
used_commission|Number|订单实际使用的佣金
used_feature_fee|Number|订单实际使用的功能费
rebate_amount|Number|订单的返佣金额
rebate_referee_addr|String|订单返佣地址
left_stock|Number|订单剩余未成交的stock数量
remain_amount|Number|订单剩余未成交的资金
deal_stock|Number|订单成交的stock数量
deal_money|Number|订单成交的money数量

### 订阅订单深度信息

字段 | 类型|描述
---|---|---
p |String|价格
a |String|数量

### 订阅交易对K线信息

字段 | 类型|描述
---|---|---
open |String| 开盘价
close|String| 收盘价
high|String|最高价
low|String|最低价
total|String|总的成交量
unix_time|Number|推送K线的时间戳
time_span|String|K线的时间范围；1min、1hour、1day
market|String| 交易对

### 订阅bancor 信息

字段 | 类型|描述
---|---|---
owner|String|bancor 的创建者
stock|String| stock token
money|String| money token
init_price|String| bancor交易对的初始化价格
max_supply|String| bancor的最大供应量
max_price|String| bancor的最大价格
current_price|String|当前bancor的价格 
stock_in_pool|String| bancor中存储的stock数量
money_in_pool|String| bancor中存储的money数量
earliest_cancel_time|Number| bancor允许取消的最早时间
stock_precision|String| stock的精度
max_money|String| bancor中存入的最大money数量
ar|String|supply-price系数

### 订阅 bancor-trade 信息

字段 | 类型|描述
---|---|---
sender|String|订单创建者
stock|String|stock token
money|String|money token
amount|Number| 订单数量
side|Number| 订单方向；BUY:1, SELL:2
money_limit|Number| 买单时愿意付出的money的上限，卖单时希望得到的money的下限
transaction_price|String| 本次交易的成交均价，单位money per stock
block_height|Number| 该笔成交所在高度
tx_hash|String|交易哈希
used_commission|Number| 订单实际使用的佣金
rebate_amount|Number| 订单的返佣金额
rebate_referee_addr|String|订单的返佣地址

### 订阅comment 信息

字段 | 类型|描述
---|---|---
id | Number|
height |Number| comment块高度 
sender|String| comment发送者
token|String| token
donation|Number| 该comment打赏的金额
title|String| title
content|String| 内容
content_type|Number| 内容类型
references|[]Object(CommentRef)|
tx_hash|String| 交易哈希

**CommentRef**

字段 | 类型|描述
---|---|---
id | Number|
reward_target|String|
reward_token|String|
reward_amount|Number|
attitudes|[]Number|

### 订阅unlock 信息

字段 | 类型|描述
---|---|---
address| String|用户地址
unlocked| []Object(sdk.Coin)|账户解锁金额
locked_coins|[]Object(LockedCoin)|账户剩余的锁定金额
frozen_coins|[]Object(sdk.Coin)|账户当前的冻结金额
coins|[]Object(sdk.Coin)|账户的可用金额
height|Number|当前块高度

**LockedCoin**

字段 | 类型|描述
---|---|---
coin | Object(sdk.Coin)| 金额
unlock_time|Number|解锁时间
from_address|String|发起锁定的地址
supervisor|String|管理员地址
reward|Number|给supervisor的奖励金额

**sdk.Coin**

字段 | 类型|描述
---|---|---
denom |String|token 符号
amount|String|token 数量

### 订阅 redelegation 信息

字段 | 类型|描述
---|---|---
delegator|String|委托者地址
src|String|前一个validator地址
dst|String|当前的validator地址
amount|String|金额
completion_time|Number|完成时间 
tx_hash|String| 交易哈希

### 订阅 unbonding 信息

字段 | 类型|描述
---|---|---
delegator|String|委托者地址
validator|String|validator地址
amount|String|金额
completion_time|Number|完成时间
tx_hash|String| 交易哈希

### 订阅slash 信息

字段 | 类型|描述
---|---|---
validator|String|validator地址
power|String|validator的voting powe
reason|String|slash原因
jailed|Boolean|是否被slash
