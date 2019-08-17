package core

import (
	//	"github.com/coinexchain/dex/app"
	//	"github.com/coinexchain/dex/modules/authx"
	//	"github.com/coinexchain/dex/modules/bancorlite"
	//	"github.com/coinexchain/dex/modules/comment"
	//	"github.com/coinexchain/dex/modules/market"

	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	cmn "github.com/tendermint/tendermint/libs/common"
)

//type (
//	NewHeightInfo                    = market.NewHeightInfo
//	CreateOrderInfo                  = market.CreateOrderInfo
//	FillOrderInfo                    = market.FillOrderInfo
//	CancelOrderInfo                  = market.CancelOrderInfo
//	NotificationSlash                = app.NotificationSlash
//	TransferRecord                   = app.TransferRecord
//	NotificationTx                   = app.NotificationTx
//	NotificationBeginRedelegation    = app.NotificationBeginRedelegation
//	NotificationBeginUnbonding       = app.NotificationBeginUnbonding
//	NotificationCompleteRedelegation = app.NotificationCompleteRedelegation
//	NotificationCompleteUnbonding    = app.NotificationCompleteUnbonding
//	NotificationUnlock               = authx.NotificationUnlock
//	TokenComment                     = comment.TokenComment
//	CommentRef                       = comment.CommentRef
//	MsgBancorTradeInfoForKafka       = bancorlite.MsgBancorTradeInfoForKafka
//	MsgBancorInfoForKafka            = bancorlite.MsgBancorInfoForKafka
//)
//
//var (
//	DecToBigEndianBytes = market.DecToBigEndianBytes
//)

const DecByteCount = 40
const SELL = 2

type CreateOrderInfo struct {
	OrderID     string  `json:"order_id"`
	Sender      string  `json:"sender"`
	TradingPair string  `json:"trading_pair"`
	OrderType   byte    `json:"order_type"`
	Price       sdk.Dec `json:"price"`
	Quantity    int64   `json:"quantity"`
	Side        byte    `json:"side"`
	TimeInForce int     `json:"time_in_force"`
	FeatureFee  int64   `json:"feature_fee"`
	Height      int64   `json:"height"`
	FrozenFee   int64   `json:"frozen_fee"`
	Freeze      int64   `json:"freeze"`
}

type FillOrderInfo struct {
	OrderID     string  `json:"order_id"`
	TradingPair string  `json:"trading_pair"`
	Height      int64   `json:"height"`
	Side        byte    `json:"side"`
	Price       sdk.Dec `json:"price"`

	// These fields will change when order was filled/canceled.
	LeftStock int64 `json:"left_stock"`
	Freeze    int64 `json:"freeze"`
	DealStock int64 `json:"deal_stock"`
	DealMoney int64 `json:"deal_money"`
	CurrStock int64 `json:"curr_stock"`
	CurrMoney int64 `json:"curr_money"`
}

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
}

type NewHeightInfo struct {
	Height        int64        `json:"height"`
	TimeStamp     time.Time    `json:"timestamp"`
	LastBlockHash cmn.HexBytes `json:"last_block_hash"`
}

type TransferRecord struct {
	Sender    string `json:"sender"`
	Recipient string `json:"recipient"`
	Amount    string `json:"amount"`
}

type NotificationTx struct {
	Signers      []sdk.AccAddress `json:"signers"`
	Transfers    []TransferRecord `json:"transfers"`
	SerialNumber int64            `json:"serial_number"`
	MsgTypes     []string         `json:"msg_types"`
	TxJSON       string           `json:"tx_json"`
	Height       int64            `json:"height"`
}

type NotificationBeginRedelegation struct {
	Delegator      string `json:"delegator"`
	ValidatorSrc   string `json:"src"`
	ValidatorDst   string `json:"dst"`
	Amount         string `json:"amount"`
	CompletionTime string `json:"completion_time"`
}

type NotificationBeginUnbonding struct {
	Delegator      string `json:"delegator"`
	Validator      string `json:"validator"`
	Amount         string `json:"amount"`
	CompletionTime string `json:"completion_time"`
}

type NotificationCompleteRedelegation struct {
	Delegator    string `json:"delegator"`
	ValidatorSrc string `json:"src"`
	ValidatorDst string `json:"dst"`
}

type NotificationCompleteUnbonding struct {
	Delegator string `json:"delegator"`
	Validator string `json:"validator"`
}

type NotificationSlash struct {
	Validator string `json:"validator"`
	Power     string `json:"power"`
	Reason    string `json:"reason"`
	Jailed    bool   `json:"jailed"`
}

type LockedCoin struct {
	Coin       sdk.Coin `json:"coin"`
	UnlockTime int64    `json:"unlock_time"`
}
type LockedCoins []LockedCoin

type NotificationUnlock struct {
	Address     sdk.AccAddress `json:"address" yaml:"address"`
	Unlocked    sdk.Coins      `json:"unlocked"`
	LockedCoins LockedCoins    `json:"locked_coins"`
	FrozenCoins sdk.Coins      `json:"frozen_coins"`
	Coins       sdk.Coins      `json:"coins" yaml:"coins"`
	Height      int64          `json:"height"`
}

type CommentRef struct {
	ID           uint64         `json:"id"`
	RewardTarget sdk.AccAddress `json:"reward_target"`
	RewardToken  string         `json:"reward_token"`
	RewardAmount int64          `json:"reward_amount"`
	Attitudes    []int32        `json:"attitudes"`
}

type TokenComment struct {
	ID          uint64         `json:"id"`
	Height      int64          `json:"height"`
	Sender      sdk.AccAddress `json:"sender"`
	Token       string         `json:"token"`
	Donation    int64          `json:"donation"`
	Title       string         `json:"title"`
	Content     string         `json:"content"`
	ContentType int8           `json:"content_type"`
	References  []CommentRef   `json:"references"`
}

type MsgBancorInfoForKafka struct {
	Owner              sdk.AccAddress `json:"sender"`
	Stock              string         `json:"stock"`
	Money              string         `json:"money"`
	InitPrice          sdk.Dec        `json:"init_price"`
	MaxSupply          sdk.Int        `json:"max_supply"`
	MaxPrice           sdk.Dec        `json:"max_price"`
	Price              sdk.Dec        `json:"price"`
	StockInPool        sdk.Int        `json:"stock_in_pool"`
	MoneyInPool        sdk.Int        `json:"money_in_pool"`
	EarliestCancelTime int64          `json:"earliest_cancel_time"`
	BlockHeight        int64          `json:"block_height"`
}

type MsgBancorTradeInfoForKafka struct {
	Sender      sdk.AccAddress `json:"sender"`
	Stock       string         `json:"stock"`
	Money       string         `json:"money"`
	Amount      int64          `json:"amount"`
	Side        byte           `json:"side"`
	MoneyLimit  int64          `json:"money_limit"`
	TxPrice     sdk.Dec        `json:"transaction_price"`
	BlockHeight int64          `json:"block_height"`
}

func DecToBigEndianBytes(d sdk.Dec) []byte {
	var result [DecByteCount]byte
	bytes := d.Int.Bytes() //  returns the absolute value of d as a big-endian byte slice.
	for i := 1; i <= len(bytes); i++ {
		result[DecByteCount-i] = bytes[len(bytes)-i]
	}
	return result[:]
}
