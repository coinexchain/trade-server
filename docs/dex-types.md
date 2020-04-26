# Dex Types数据描述

本文档描述的是 `cetd --> trade-server` 之间消息的数据结构，定义在[core/dex_type.go](../core/dex_types.go)文件内.

## Market Information 跟市场/订单有关的数据

### MarketInfo  市场信息
字段 | 类型 | 描述
---|---|---
Stock | string | 被交易的币种
Money | string | 交易金额币种
Creator | string | 创建人
PricePrecision | byte | 价格精确度（小数位，控制价格的小数位） 
OrderPrecision | byte | 订单精确度（控制每笔订单中交易token的数量，token交易的数量必须是(`10**OrderPrecision`)的倍数）

### OrderInfo  订单信息
字段 | 类型 | 描述
---|---|---
CreateOrderInfo | OrderResponse | 创建订单信息
FillOrderInfo | OrderResponse | 成交订单信息
CancelOrderInfo | OrderResponse | 取消订单信息

### OrderResponse  订单信息返回值
字段 | 类型 | 描述
---|---|---
Data | []json.RawMessage | 订单信息（创建/成交/取消）
Timesid | []int64 | index 0: time(该Msg的区块时间戳)； index 1: sid(消息的累计序列号，在trade-server中作为数据库中key的一部分，用于范围查询) 

### CreateOrderInfo  创建订单信息
字段 | 类型 | 描述
---|---|---
OrderID | string | 订单ID
Sender | string | 创建者
TradingPair | string | 交易对（例如 abc/cet）
OrderType | byte | 2:限价单; 当前只支持限价单。
Price | sdk.Dec | 挂单价
Quantity | int64 | 交易stock的数量
Side | byte | 订单方向；BUY:1, SELL:2
TimeInForce | int | 订单类型；GTE:到期单, IOC:立即成交单
Height | int64 | 创建订单的高度
Freeze | int64 | 订单使用的资金
FrozenFeatureFee | int64 | 冻结订单的最大功能费；订单删除时，按存在时间返还
FrozenCommission | int64 | 冻结订单的最大佣金；订单删除时，按成交情况返还
TxHash | string | 交易哈希

### FillOrderInfo  成交订单信息
字段 | 类型 | 描述
---|---|---
OrderID | string | 订单ID
TradingPair | string | 交易对（例如 abc/cet）
Height | int64 | 订单成交的高度
Side | byte |  订单方向；BUY:1, SELL:2
Price | sdk.Dec | 挂单价
LeftStock | int64 | 订单未成交的stock
Freeze | int64 | 订单剩余还未成交的冻结资金
DealStock | int64 | 订单成交的stock
DealMoney | int64 | 订单成交的money
CurrStock | int64 | 订单本次成交的stock数量
CurrMoney | int64 | 订单本次成交的money数量
FillPrice | sdk.Dec | 交易哈希

### CancelOrderInfo  取消订单信息
字段 | 类型 | 描述
---|---|---
OrderID | string | 订单ID
TradingPair | string | 交易对（例如 abc/cet）
Height | int64 | 删除订单时的块高度
Side | byte |  订单方向；BUY:1, SELL:2
Price | sdk.Dec | 挂单价
DelReason | string | 订单取消原因
UsedCommission | int64 | 订单实际使用的佣金
LeftStock | int64 | 订单剩余未成交的stock数量
RemainAmount | int64 | 订单剩余未成交的资金
DealStock | int64 | 订单成交的stock数量
DealMoney | int64 | 订单成交的money数量
TxHash | string | 交易哈希
UsedFeatureFee | int64 | 订单实际使用的功能费
RebateAmount | int64 | 订单的返佣金额
RebateRefereeAddr | string | 订单返佣地址

### DepthDetails 交易市场深度
字段 | 类型 | 描述
---|---|---
TradingPair | string | 交易对（例如 abc/cet）
Bids | []*PricePoint | 买家报价-数量列表
Asks | []*PricePoint | 卖家报价-数量列表


## Block Information 跟区块本身有关的数据

### NewHeightInfo  区块高度信息

对比起旧链，新链的timestamp由string改成int64

字段 | 类型 | 描述
---|---|---
ChainID | string | 链ID
Height | int64 | 区块高度
TimeStamp | int64 | 时间戳
LastBlockHash | cmn.HexBytes | 上一个区块的hash

### HeightInfo  高度信息
字段 | 类型 | 描述
---|---|---
Height | uint64 | 区块高度
TimeStamp | uint64 | 时间戳

### TransferRecord  转账记录
字段 | 类型 | 描述
---|---|---
Sender | string | 发送人
Recipient | string | 接收者
Amount | string | 数量

### NotificationTx 交易信息
字段 | 类型 | 描述
---|---|---
Signers | []string | 签名者
Transfers | []TransferRecord | 转账记录列表
SerialNumber | int64 | 当前交易的序号
MsgTypes | []string | 交易内的消息类型
TxJSON | string | 交易JSON格式字符串
Height | int64 | 交易发生高度
Hash | string | 交易哈希
ExtraInfo | string | 交易失败时，提供的额外信息


## Management Information 跟共识/治理有关的数据

### NotificationBeginRedelegation

字段 | 类型 | 描述
---|---|---
Delegator | string | 委托人
ValidatorSrc | string | 原来的被委托者
ValidatorDst | string | 新的被委托者
Amount | string | 委托的money
CompletionTime | int64 | 完成时间
TxHash | string | 交易哈希

### NotificationBeginUnbonding

字段 | 类型 | 描述
---|---|---
Delegator | string | 委托人
Validator | string | 验证人
Amount | string | 委托的money
CompletionTime | int64 | 完成时间
TxHash | string | 交易哈希

### NotificationCompleteRedelegation

BeginRedelegation完成信息

字段 | 类型 | 描述
---|---|---
Delegator | string | 委托人
ValidatorSrc | string | 原来的被委托者
ValidatorDst | string | 新的被委托者

### NotificationCompleteUnbonding

BeginUnbonding完成信息

字段 | 类型 | 描述
---|---|---
Delegator | string | 委托人
Validator | string | 验证人

### NotificationSlash

当一个validator做出违规行为(活性差，双签)，他会被惩罚，票数会削减一定百分比

字段 | 类型 | 描述
---|---|---
Validator | string | 验证人
Power | string | 票数削减百分比
Reason | string | 理由
Jailed | bool | 是否被监禁惩罚

## Bancor Information 跟Bancor有关的数据

### MsgBancorInfoForKafka

使用bancor框架自动做市

字段 | 类型 | 描述
---|---|---
Owner | string | bancor的创建者
Stock | string | 被交易的币种
Money | string | 交易金额币种
InitPrice | string | bancor交易对的初始化价格，创建后币价达到这个值
MaxSupply | string | bancor的最大供应量
StockPrecision | string | stock的精度
MaxPrice | string | bancor的最大价格
MaxMoney | string | bancor中存入的最大money数量
AR | string | supply-price系数
CurrentPrice | string | 当前bancor的价格，CurrentPrice =  (MaxSupply - StockInPool) * MaxPrice / MaxSupply
StockInPool | string | bancor中存储的stock数量
MoneyInPool | string | bancor中存储的money数量，MoneyInPool = CurrentPrice * (MaxSupply - StockInPool) * 0.5
EarliestCancelTime | int64 | bancor允许取消的最早时间，只有onwer才能取消

### MsgBancorTradeInfoForKafka

bancor交易信息

字段 | 类型 | 描述
---|---|---
Sender | string | 交易人
Stock | string | 被交易的币种
Money | string | 交易金额币种
Amount | int64 | 订单数量
Side | byte | 订单方向；BUY:1, SELL:2
MoneyLimit | int64 | 买单时愿意付出的money的上限，卖单时希望得到的money的下限
TxPrice | sdk.Dec | 本次交易的成交均价，单位money per stock
BlockHeight | int64 | 该笔成交所在高度
TxHash | string | 交易哈希
UsedCommission | int64 | 订单实际使用的佣金
RebateAmount | int64 | 订单的返佣金额
RebateRefereeAddr | string | 订单的返佣地址


## Reward Information 跟奖励有关的数据

### NotificationValidatorCommission

validator收到的佣金

字段 | 类型 | 描述
---|---|---
Validator | string | 验证人
Commission | string | 佣金

### NotificationDelegatorRewards

该validator的投票者收到的奖励

字段 | 类型 | 描述
---|---|---
Validator | string | 验证人
Rewards | string | 奖励

## Other Information 其他数据

### LockedCoin  锁定金额

字段 | 类型 | 描述
---|---|---
Coin | sdk.Coin | 金额（sdk.Coin包含货币描述和数量）
UnlockTime | int64 | 解锁时间
FromAddress | string | 发起锁定的地址
Supervisor | string | 管理员地址
Reward | int64 | 管理员的奖励

### NotificationUnlock  解锁信息
字段 | 类型 | 描述
---|---|---
Address | string | 用户地址
Unlocked | sdk.Coins | 账户解锁金额
LockedCoins | []LockedCoin | 账户剩余的锁定金额
FrozenCoins | sdk.Coins | 账户当前的冻结金额
Coins | sdk.Coins | 账户的可用金额
Height | int64 | 当前区块高度

### LockedSendMsg  发送锁定货币

A给B发送锁定货币，B只有等到解锁时间过后才能使用

字段 | 类型 | 描述
---|---|---
FromAddress | string | 发送人
ToAddress | string | 接收人
Amount | sdk.Coins | 货币信息
UnlockTime | int64 | 解锁时间
Supervisor | string | 管理员
Reward | int64 | 管理员的奖励
TxHash | string | 交易哈希

### CommentRef
字段 | 类型 | 描述
---|---|---
ID | uint64 | Comment信息ID
RewardTarget | string | 奖励地址
RewardToken | string | 奖励币种
RewardAmount | int64 | 奖励金额
Attitudes | []int32 | 对该评论的态度(like,Dislike,Laugh...; 11种态度)

### TokenComment
字段 | 类型 | 描述
---|---|---
ID | uint64 | 评论指定token时，该评论的唯一ID；由系统维护，单调递增;
Height | int64 | comment所在区块高度
Sender | string | comment发送者
Token | string | 打赏币种
Donation | int64 | 该comment打赏的金额
Title | string | 标题
Content | string | 内容
ContentType | int8 | 内容类型
References | []CommentRef | 奖励
TxHash | string | 交易哈希
