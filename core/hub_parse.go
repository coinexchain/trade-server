package core

import (
	"encoding/json"
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"
	cmn "github.com/tendermint/tendermint/libs/common"
)

type OldHeightInfo struct {
	ChainID       string       `json:"chain_id,omitempty"`
	Height        int64        `json:"height"`
	TimeStamp     string       `json:"timestamp"`
	LastBlockHash cmn.HexBytes `json:"last_block_hash"`
}

func (hub *Hub) parseHeightInfo(bz []byte, convert bool) *NewHeightInfo {
	if convert {
		return convertNewHeightInfo(bz)
	}
	var v NewHeightInfo
	if err := json.Unmarshal(bz, &v); err != nil {
		log.Error("Error in Unmarshal NewHeightInfo", err)
		return nil
	}
	return &v
}

func convertNewHeightInfo(bz []byte) *NewHeightInfo {
	var tmp OldHeightInfo
	if err := json.Unmarshal(bz, &tmp); err != nil {
		log.Error("Parse error in Unmarshal NewHeightInfo : ", err)
		return nil
	}
	return tmp.convertToNewHeightInfo()
}

func (o *OldHeightInfo) convertToNewHeightInfo() *NewHeightInfo {
	t, err := time.Parse(time.RFC3339, o.TimeStamp)
	if err != nil {
		log.Error(fmt.Sprintf("parse time error; data [%s]", o.TimeStamp))
		return nil
	}
	return &NewHeightInfo{
		ChainID:       o.ChainID,
		Height:        o.Height,
		TimeStamp:     t.Unix(),
		LastBlockHash: o.LastBlockHash,
	}
}

type OldBancorInfoForKafka struct {
	Owner              string `json:"sender"`
	Stock              string `json:"stock"`
	Money              string `json:"money"`
	InitPrice          string `json:"init_price"`
	MaxSupply          string `json:"max_supply"`
	StockPrecision     byte   `json:"stock_precision"`
	MaxPrice           string `json:"max_price"`
	Price              string `json:"price"`
	StockInPool        string `json:"stock_in_pool"`
	MoneyInPool        string `json:"money_in_pool"`
	EarliestCancelTime int64  `json:"earliest_cancel_time"`
}

func (hub *Hub) parseBancorInfo(bz []byte) *MsgBancorInfoForKafka {
	if hub.chainID == hub.oldChainID {
		return convertNewMsgBancorInfo(bz)
	}
	var v MsgBancorInfoForKafka
	if err := json.Unmarshal(bz, &v); err != nil {
		log.Error("Parse error in unmarshal MsgBancorInfoForKafka : ", err)
		return nil
	}
	return &v
}

func convertNewMsgBancorInfo(bz []byte) *MsgBancorInfoForKafka {
	var old OldBancorInfoForKafka
	if err := json.Unmarshal(bz, &old); err != nil {
		log.Error("Parse error in unmarshal OldBancorInfoForKafka : ", err)
		return nil
	}
	return old.convertToNewBancorInfo()
}

func (o *OldBancorInfoForKafka) convertToNewBancorInfo() *MsgBancorInfoForKafka {
	return &MsgBancorInfoForKafka{
		Owner:              o.Owner,
		Stock:              o.Stock,
		Money:              o.Money,
		InitPrice:          o.InitPrice,
		MaxSupply:          o.MaxSupply,
		StockPrecision:     fmt.Sprintf("%d", o.StockPrecision),
		MaxPrice:           o.MaxPrice,
		MaxMoney:           "",
		AR:                 "",
		CurrentPrice:       o.Price,
		StockInPool:        o.StockInPool,
		MoneyInPool:        o.MoneyInPool,
		EarliestCancelTime: o.EarliestCancelTime,
	}
}

type OldNotificationBeginRedelegation struct {
	Delegator      string `json:"delegator"`
	ValidatorSrc   string `json:"src"`
	ValidatorDst   string `json:"dst"`
	Amount         string `json:"amount"`
	CompletionTime string `json:"completion_time"`
}

func (hub *Hub) parseNotificationBeginRedelegation(bz []byte) *NotificationBeginRedelegation {
	if hub.chainID == hub.oldChainID {
		return convertNewNotificationBeginRedelegation(bz)
	}
	var v NotificationBeginRedelegation
	if err := json.Unmarshal(bz, &v); err != nil {
		log.Error("Parse error in unmarshal NotificationBeginRedelegation : ", err)
		return nil
	}
	return &v
}

func convertNewNotificationBeginRedelegation(bz []byte) *NotificationBeginRedelegation {
	var old OldNotificationBeginRedelegation
	if err := json.Unmarshal(bz, &old); err != nil {
		log.Error("Parse error in unmarshal OldNotificationBeginRedelegation : ", err)
		return nil
	}
	return old.convertToNewNotificationBeginRedelegation()
}

func (old *OldNotificationBeginRedelegation) convertToNewNotificationBeginRedelegation() *NotificationBeginRedelegation {
	t, err := time.Parse(time.RFC3339, old.CompletionTime)
	if err != nil {
		log.Error(fmt.Sprintf("parse time error; data [%s]", old.CompletionTime))
		return nil
	}
	return &NotificationBeginRedelegation{
		Delegator:      old.Delegator,
		ValidatorSrc:   old.ValidatorSrc,
		ValidatorDst:   old.ValidatorDst,
		Amount:         old.Amount,
		CompletionTime: t.Unix(),
	}
}

type OldNotificationBeginUnbonding struct {
	Delegator      string `json:"delegator"`
	Validator      string `json:"validator"`
	Amount         string `json:"amount"`
	CompletionTime string `json:"completion_time"`
}

func (hub *Hub) parseNotificationBeginUnbonding(bz []byte) *NotificationBeginUnbonding {
	if hub.chainID == hub.oldChainID {
		return convertNewNotificationBeginUnbonding(bz)
	}
	var v NotificationBeginUnbonding
	if err := json.Unmarshal(bz, &v); err != nil {
		log.Error("Parse error in unmarshal NotificationBeginUnbonding : ", err)
		return nil
	}
	return &v
}

func convertNewNotificationBeginUnbonding(bz []byte) *NotificationBeginUnbonding {
	var old OldNotificationBeginUnbonding
	if err := json.Unmarshal(bz, &old); err != nil {
		log.Error("Parse error in unmarshal OldNotificationBeginUnbonding : ", err)
		return nil
	}
	return old.convertToNewNotificationBeginUnbonding()
}

func (old *OldNotificationBeginUnbonding) convertToNewNotificationBeginUnbonding() *NotificationBeginUnbonding {
	t, err := time.Parse(time.RFC3339, old.CompletionTime)
	if err != nil {
		log.Error(fmt.Sprintf("parse time error; data [%s]", old.CompletionTime))
		return nil
	}
	return &NotificationBeginUnbonding{
		Delegator:      old.Delegator,
		Validator:      old.Validator,
		Amount:         old.Amount,
		CompletionTime: t.Unix(),
	}
}
