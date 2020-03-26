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

func (hub *Hub) parseHeightInfo(bz []byte, convert bool) *NewHeightInfo {
	if convert {
		return hub.convertNewHeightInfo(bz)
	}
	var v NewHeightInfo
	if err := json.Unmarshal(bz, &v); err != nil {
		log.Error("Error in Unmarshal NewHeightInfo", err)
		return nil
	}
	return &v
}

func (hub *Hub) convertNewHeightInfo(bz []byte) *NewHeightInfo {
	var tmp OldHeightInfo
	err := json.Unmarshal(bz, &tmp)
	if err != nil {
		log.Error("Error in Unmarshal NewHeightInfo")
		return nil
	}
	return tmp.convertToNewHeightInfo()
}
