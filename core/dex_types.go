package core

import (
	//	"github.com/coinexchain/dex/app"
	//	"github.com/coinexchain/dex/modules/authx"
	//	"github.com/coinexchain/dex/modules/bancorlite"
	//	"github.com/coinexchain/dex/modules/comment"
	//	"github.com/coinexchain/dex/modules/market"

	"encoding/json"

	sdk "github.com/cosmos/cosmos-sdk/types"
	cmn "github.com/tendermint/tendermint/libs/common"
)

const (
	DecByteCount = 40
	BUY          = 1
	SELL         = 2
	GTE          = 3
	LIMIT        = 2
)

// --------------------------
// market info, e.g. abc/cet
type MarketInfo struct {
	Stock          string `json:"stock"`
	Money          string `json:"money"`
	Creator        string `json:"creator"`
	PricePrecision byte   `json:"price_precision"`
	OrderPrecision byte   `json:"order_precision"`
}

type OrderInfo struct {
	CreateOrderInfo OrderResponse `json:"create_order_info"`
	FillOrderInfo   OrderResponse `json:"fill_order_info"`
	CancelOrderInfo OrderResponse `json:"cancel_order_info"`
}

type OrderResponse struct {
	Data    []json.RawMessage `json:"data"`
	Timesid []int64           `json:"timesid"`
}

// someone creates an order
type CreateOrderInfo struct {
	OrderID     string  `json:"order_id"`
	Sender      string  `json:"sender"`
	TradingPair string  `json:"trading_pair"`
	OrderType   byte    `json:"order_type"`
	Price       sdk.Dec `json:"price"`
	Quantity    int64   `json:"quantity"`
	Side        byte    `json:"side"`
	TimeInForce int     `json:"time_in_force"`
	FeatureFee  int64   `json:"feature_fee"` //Deprecated
	Height      int64   `json:"height"`
	FrozenFee   int64   `json:"frozen_fee"` //Deprecated
	Freeze      int64   `json:"freeze"`

	FrozenFeatureFee int64 `json:"frozen_feature_fee,omitempty"`
	FrozenCommission int64 `json:"frozen_commission,omitempty"`

	TxHash string `json:"tx_hash,omitempty"`
}

// another guy fills someone's order
type FillOrderInfo struct {
	OrderID     string  `json:"order_id"`
	TradingPair string  `json:"trading_pair"`
	Height      int64   `json:"height"`
	Side        byte    `json:"side"`
	Price       sdk.Dec `json:"price"`

	// These fields will change when order was filled/canceled.
	LeftStock int64   `json:"left_stock"`
	Freeze    int64   `json:"freeze"`
	DealStock int64   `json:"deal_stock"`
	DealMoney int64   `json:"deal_money"`
	CurrStock int64   `json:"curr_stock"`
	CurrMoney int64   `json:"curr_money"`
	FillPrice sdk.Dec `json:"fill_price"`
}

// someone who has created an order before, but now he/she wants to cancel it
type CancelOrderInfo struct {
	OrderID     string  `json:"order_id"`
	TradingPair string  `json:"trading_pair"`
	Height      int64   `json:"height"`
	Side        byte    `json:"side"`
	Price       sdk.Dec `json:"price"`

	// Del infos
	DelReason string `json:"del_reason"`

	// Fields of amount
	UsedCommission int64 `json:"used_commission"`
	LeftStock      int64 `json:"left_stock"`
	RemainAmount   int64 `json:"remain_amount"`
	DealStock      int64 `json:"deal_stock"`
	DealMoney      int64 `json:"deal_money"`

	TxHash string `json:"tx_hash,omitempty"`

	UsedFeatureFee    int64  `json:"used_feature_fee,omitempty"`
	RebateAmount      int64  `json:"rebate_amount,omitempty"`
	RebateRefereeAddr string `json:"rebate_referee_addr,omitempty"`
}

// for the same trading pair, the bid prices and ask prices in this market
type DepthDetails struct {
	TradingPair string        `json:"trading_pair"`
	Bids        []*PricePoint `json:"bids"`
	Asks        []*PricePoint `json:"asks"`
}

// ---------------------
// common info
// timestamp field changes from string to int64
type NewHeightInfo struct {
	ChainID       string       `json:"chain_id,omitempty"`
	Height        int64        `json:"height"`
	TimeStamp     int64        `json:"timestamp"`
	LastBlockHash cmn.HexBytes `json:"last_block_hash"`
}

type HeightInfo struct {
	Height    uint64 `json:"height"`
	TimeStamp uint64 `json:"timestamp"`
}

type TransferRecord struct {
	Sender    string `json:"sender"`
	Recipient string `json:"recipient"`
	Amount    string `json:"amount"`
}

// transaction
type NotificationTx struct {
	Signers      []string         `json:"signers"`
	Transfers    []TransferRecord `json:"transfers"`
	SerialNumber int64            `json:"serial_number"`
	MsgTypes     []string         `json:"msg_types"`
	TxJSON       string           `json:"tx_json"`
	Height       int64            `json:"height"`
	Hash         string           `json:"hash"`
	ExtraInfo    string           `json:"extra_info,omitempty"`
}

// ------------------------
// manager info
// someone is going to delegate a validator to stake
// or, someone who voted a validator before, but now he/she wants to delegate another validator
type NotificationBeginRedelegation struct {
	Delegator      string `json:"delegator"`
	ValidatorSrc   string `json:"src"`
	ValidatorDst   string `json:"dst"`
	Amount         string `json:"amount"`
	CompletionTime int64  `json:"completion_time"`

	TxHash string `json:"tx_hash,omitempty"`
}

// a validator leaves top 42
// or a jailed validator becomes unjailed
type NotificationBeginUnbonding struct {
	Delegator      string `json:"delegator"`
	Validator      string `json:"validator"`
	Amount         string `json:"amount"`
	CompletionTime int64  `json:"completion_time"`

	TxHash string `json:"tx_hash,omitempty"`
}

// the begin redelegation process is completed
type NotificationCompleteRedelegation struct {
	Delegator    string `json:"delegator"`
	ValidatorSrc string `json:"src"`
	ValidatorDst string `json:"dst"`
}

// the unbounding process is completed
type NotificationCompleteUnbonding struct {
	Delegator string `json:"delegator"`
	Validator string `json:"validator"`
}

// if validators do some illegal acts，they will be punished (e.g. losing votes) or even jailed
type NotificationSlash struct {
	Validator string `json:"validator"`
	Power     string `json:"power"`
	Reason    string `json:"reason"`
	Jailed    bool   `json:"jailed"`
}

// -----------
// bancor info
// someone uses bancor to create markets automatically
type MsgBancorInfoForKafka struct {
	Owner              string `json:"owner"`
	Stock              string `json:"stock"`
	Money              string `json:"money"`
	InitPrice          string `json:"init_price"`
	MaxSupply          string `json:"max_supply"`
	StockPrecision     string `json:"stock_precision,omitempty"`
	MaxPrice           string `json:"max_price"`
	MaxMoney           string `json:"max_money,omitempty"`
	AR                 string `json:"ar,omitempty"`
	CurrentPrice       string `json:"current_price"`
	StockInPool        string `json:"stock_in_pool"`
	MoneyInPool        string `json:"money_in_pool"`
	EarliestCancelTime int64  `json:"earliest_cancel_time"`
}

// some users trade with bancor
type MsgBancorTradeInfoForKafka struct {
	Sender      string  `json:"sender"`
	Stock       string  `json:"stock"`
	Money       string  `json:"money"`
	Amount      int64   `json:"amount"`
	Side        byte    `json:"side"`
	MoneyLimit  int64   `json:"money_limit"`
	TxPrice     sdk.Dec `json:"transaction_price"`
	BlockHeight int64   `json:"block_height"`

	TxHash string `json:"tx_hash,omitempty"`

	UsedCommission    int64  `json:"used_commission,omitempty"`
	RebateAmount      int64  `json:"rebate_amount,omitempty"`
	RebateRefereeAddr string `json:"rebate_referee_addr,omitempty"`
}

// -------------
// block reward info
type NotificationValidatorCommission struct {
	Validator  string `json:"validator"`
	Commission string `json:"commission"`
}
type NotificationDelegatorRewards struct {
	Validator string `json:"validator"`
	Rewards   string `json:"rewards"`
}

// --------------------
// other module info
type LockedCoin struct {
	Coin        sdk.Coin `json:"coin"`
	UnlockTime  int64    `json:"unlock_time"`
	FromAddress string   `json:"from_address,omitempty"`
	Supervisor  string   `json:"supervisor,omitempty"`
	Reward      int64    `json:"reward,omitempty"`
}
type LockedCoins []LockedCoin

// unclock the locked coins
type NotificationUnlock struct {
	Address     string      `json:"address" yaml:"address"`
	Unlocked    sdk.Coins   `json:"unlocked"`
	LockedCoins LockedCoins `json:"locked_coins"`
	FrozenCoins sdk.Coins   `json:"frozen_coins"`
	Coins       sdk.Coins   `json:"coins" yaml:"coins"`
	Height      int64       `json:"height"`
}

// someone sends locked coins to another
// another person can only wait until the unlock time to use these coins
type LockedSendMsg struct {
	FromAddress string    `json:"from_address"`
	ToAddress   string    `json:"to_address"`
	Amount      sdk.Coins `json:"amount"`
	UnlockTime  int64     `json:"unlock_time"`
	Supervisor  string    `json:"supervisor,omitempty"`
	Reward      int64     `json:"reward,omitempty"`

	TxHash string `json:"tx_hash,omitempty"`
}

type CommentRef struct {
	ID           uint64  `json:"id"`
	RewardTarget string  `json:"reward_target"`
	RewardToken  string  `json:"reward_token"`
	RewardAmount int64   `json:"reward_amount"`
	Attitudes    []int32 `json:"attitudes"`
}

type TokenComment struct {
	ID          uint64       `json:"id"`
	Height      int64        `json:"height"`
	Sender      string       `json:"sender"`
	Token       string       `json:"token"`
	Donation    int64        `json:"donation"`
	Title       string       `json:"title"`
	Content     string       `json:"content"`
	ContentType int8         `json:"content_type"`
	References  []CommentRef `json:"references"`

	TxHash string `json:"tx_hash,omitempty"`
}
