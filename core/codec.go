package core

import (
	"encoding/json"

	log "github.com/sirupsen/logrus"
)

type Depth struct {
	Type string        `json:"type"`
	Bids []*PricePoint `json:"bids"`
	Asks []*PricePoint `json:"asks"`
}

func encodeDepth(depth map[*PricePoint]bool, buy bool) []byte {
	values := make([]*PricePoint, 0, len(depth))
	for p := range depth {
		values = append(values, p)
	}
	msg := Depth{Type: DepthKey}
	if buy {
		msg.Bids = values
	} else {
		msg.Asks = values
	}

	bz, err := json.Marshal(msg)
	if err != nil {
		log.Error(err)
	}
	return bz
}
