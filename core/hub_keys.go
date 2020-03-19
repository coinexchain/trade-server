package core

import "time"

const (
	MaxCount     = 1024
	DumpVersion  = byte(0)
	DumpInterval = 1000
	DumpMinTime  = 10 * time.Minute
	//These bytes are used as the first byte in key
	//The format of different kinds of keys are listed below:
	CandleStickByte = byte(0x10) //-, ([]byte(market), []byte{0, timespan}...), 0, currBlockTime, hub.sid, lastByte=0
	DealByte        = byte(0x12) //-, []byte(market), 0, currBlockTime, hub.sid, lastByte=0
	OrderByte       = byte(0x14) //-, []byte(addr), 0, currBlockTime, hub.sid, lastByte=
	//                                                          CreateOrderEndByte,FillOrderEndByte,CancelOrderEndByte
	BancorInfoByte   = byte(0x16) //-, []byte(market), 0, currBlockTime, hub.sid, lastByte=0
	BancorTradeByte  = byte(0x18) //-, []byte(addr), 0, currBlockTime, hub.sid, lastByte=0
	IncomeByte       = byte(0x1A) //-, []byte(addr), 0, currBlockTime, hub.sid, lastByte=0
	TxByte           = byte(0x1C) //-, []byte(addr), 0, currBlockTime, hub.sid, lastByte=0
	CommentByte      = byte(0x1E) //-, []byte(token), 0, currBlockTime, hub.sid, lastByte=0
	BlockHeightByte  = byte(0x20) //-, heightBytes
	DetailByte       = byte(0x22) //([]byte{DetailByte}, []byte(v.Hash), currBlockTime)
	SlashByte        = byte(0x24) //-, []byte{}, 0, currBlockTime, hub.sid, lastByte=0
	LatestHeightByte = byte(0x26) //-
	BancorDealByte   = byte(0x28) //-, []byte(market), 0, currBlockTime, hub.sid, lastByte=0
	RedelegationByte = byte(0x30) //-, []byte(token), 0, completion time, hub.sid, lastByte=0
	UnbondingByte    = byte(0x32) //-, []byte(token), 0, completion time, hub.sid, lastByte=0
	UnlockByte       = byte(0x34) //-, []byte(addr), 0, currBlockTime, hub.sid, lastByte=0
	LockedByte       = byte(0x36) //-, []byte(addr), 0, currBlockTime, hub.sid, lastByte=0
	DonationByte     = byte(0x38) //-, []byte{}, 0, currBlockTime, hub.sid, lastByte=0
	DelistByte       = byte(0x3A) //-, []byte(market), 0, currBlockTime, hub.sid, lastByte=0

	// Used to store meta information
	OffsetByte = byte(0xF0)
	DumpByte   = byte(0xF1)

	// Used to indicate query types, not used in the keys in rocksdb
	DelistsByte         = byte(0x3B)
	CreateOrderOnlyByte = byte(0x40)
	FillOrderOnlyByte   = byte(0x42)
	CancelOrderOnlyByte = byte(0x44)

	// Later add key
	CreateMarketByte        = byte(0x46)
	ValidatorCommissionByte = byte(0x48)
	DelegatorRewardsByte    = byte(0x50)
)

func (hub *Hub) getCandleStickKey(market string, timespan byte) []byte {
	bz := append([]byte(market), []byte{0, timespan}...)
	return hub.getKeyFromBytes(CandleStickByte, bz, 0)
}
func (hub *Hub) getLockedKey(addr string) []byte {
	return hub.getKeyFromBytes(LockedByte, []byte(addr), 0)
}
func (hub *Hub) getDealKey(market string) []byte {
	return hub.getKeyFromBytes(DealByte, []byte(market), 0)
}
func (hub *Hub) getBancorInfoKey(market string) []byte {
	return hub.getKeyFromBytes(BancorInfoByte, []byte(market), 0)
}
func (hub *Hub) getBancorDealKey(market string) []byte {
	return hub.getKeyFromBytes(BancorDealByte, []byte(market), byte(0))
}
func (hub *Hub) getCommentKey(token string) []byte {
	return hub.getKeyFromBytes(CommentByte, []byte(token), 0)
}
func (hub *Hub) getCreateMarketKey(market string) []byte {
	return hub.getKeyFromBytes(CreateMarketByte, []byte(market), 0)
}
func (hub *Hub) getCreateOrderKey(addr string) []byte {
	return hub.getKeyFromBytes(OrderByte, []byte(addr), CreateOrderEndByte)
}
func (hub *Hub) getFillOrderKey(addr string) []byte {
	return hub.getKeyFromBytes(OrderByte, []byte(addr), FillOrderEndByte)
}
func (hub *Hub) getCancelOrderKey(addr string) []byte {
	return hub.getKeyFromBytes(OrderByte, []byte(addr), CancelOrderEndByte)
}
func (hub *Hub) getValidatorCommissionKey(addr string) []byte {
	return hub.getKeyFromBytes(ValidatorCommissionByte, []byte(addr), 0)
}
func (hub *Hub) getDelegatorRewardsKey(addr string) []byte {
	return hub.getKeyFromBytes(DelegatorRewardsByte, []byte(addr), 0)
}
func (hub *Hub) getBancorTradeKey(addr string) []byte {
	return hub.getKeyFromBytes(BancorTradeByte, []byte(addr), byte(0))
}
func (hub *Hub) getIncomeKey(addr string) []byte {
	return hub.getKeyFromBytes(IncomeByte, []byte(addr), byte(0))
}
func (hub *Hub) getTxKey(addr string) []byte {
	return hub.getKeyFromBytes(TxByte, []byte(addr), byte(0))
}
func (hub *Hub) getRedelegationEventKey(addr string, time int64) []byte {
	return hub.getKeyFromBytesAndTime(RedelegationByte, []byte(addr), byte(0), time)
}
func (hub *Hub) getUnbondingEventKey(addr string, time int64) []byte {
	return hub.getKeyFromBytesAndTime(UnbondingByte, []byte(addr), byte(0), time)
}
func (hub *Hub) getUnlockEventKey(addr string) []byte {
	return hub.getKeyFromBytes(UnlockByte, []byte(addr), byte(0))
}

func (hub *Hub) getKeyFromBytes(firstByte byte, bz []byte, lastByte byte) []byte {
	return hub.getKeyFromBytesAndTime(firstByte, bz, lastByte, hub.currBlockTime.Unix())
}

// Following are some functions to generate keys to access the KVStore
func (hub *Hub) getKeyFromBytesAndTime(firstByte byte, bz []byte, lastByte byte, unixTime int64) []byte {
	res := make([]byte, 0, 1+1+len(bz)+1+16+1)
	res = append(res, firstByte)
	res = append(res, byte(len(bz)))
	res = append(res, bz...)
	res = append(res, byte(0))
	//the block's time at which the KV pair is generated
	res = append(res, Int64ToBigEndianBytes(unixTime)...)
	// the serial ID for a KV pair
	res = append(res, Int64ToBigEndianBytes(hub.sid)...)
	res = append(res, lastByte)
	return res
}

func (hub *Hub) getEventKeyWithSidAndTime(firstByte byte, addr string, time int64, sid int64) []byte {
	res := make([]byte, 0, 1+1+len(addr)+1+16+1)
	res = append(res, firstByte)
	res = append(res, byte(len(addr)))
	res = append(res, []byte(addr)...)
	res = append(res, byte(0))
	res = append(res, Int64ToBigEndianBytes(time)...) //the block's time at which the KV pair is generated
	res = append(res, Int64ToBigEndianBytes(sid)...)  // the serial ID for a KV pair
	res = append(res, 0)
	return res
}
