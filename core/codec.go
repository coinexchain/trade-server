package core

import (
	"encoding/json"

	log "github.com/sirupsen/logrus"
)

type DepthDetails struct {
	Bids []*PricePoint `json:"bids"`
	Asks []*PricePoint `json:"asks"`
}

type Depth struct {
	Type    string       `json:"type"`
	Payload DepthDetails `json:"payload"`
}

func encodeDepth(depth map[*PricePoint]bool, buy bool) []byte {
	values := make([]*PricePoint, 0, len(depth))
	for p := range depth {
		values = append(values, p)
	}
	msg := Depth{Type: DepthKey}
	if buy {
		msg.Payload.Bids = values
	} else {
		msg.Payload.Asks = values
	}

	bz, err := json.Marshal(msg)
	if err != nil {
		log.Error(err)
	}
	return bz
}
