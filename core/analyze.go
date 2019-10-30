package core

// Here are some analysis functions, used for debug

import (
	"encoding/json"
	"math"
)

type AllOrderInfo struct {
	CreateMap map[string]CreateOrderInfo
	FillMap   map[string]FillOrderInfo
	CancelMap map[string]CancelOrderInfo
}

func (hub *Hub) GetAllOrderInfo() AllOrderInfo {
	res := AllOrderInfo{
		CreateMap: make(map[string]CreateOrderInfo),
		FillMap:   make(map[string]FillOrderInfo),
		CancelMap: make(map[string]CancelOrderInfo),
	}
	var vCreate CreateOrderInfo
	var vFill FillOrderInfo
	var vCancel CancelOrderInfo
	data, tags, timesid := hub.query(false, OrderByte, []byte{}, math.MaxInt64, 0, MaxCount, nil)
	for len(data) < MaxCount {
		for i, tag := range tags {
			if tag == CreateOrderEndByte {
				json.Unmarshal(data[i], &vCreate)
				res.CreateMap[vCreate.OrderID] = vCreate
			} else if tag == FillOrderEndByte {
				json.Unmarshal(data[i], &vFill)
				res.FillMap[vCreate.OrderID] = vFill
			} else if tag == CancelOrderEndByte {
				json.Unmarshal(data[i], &vCancel)
				res.CancelMap[vCreate.OrderID] = vCancel
			} else {
				panic("Unknown tag.")
			}
		}
		lastSid := timesid[len(timesid)-1]
		lastTime := timesid[len(timesid)-2]
		data, tags, timesid = hub.query(false, OrderByte, []byte{}, lastTime, lastSid, MaxCount, nil)
	}
	return res
}
