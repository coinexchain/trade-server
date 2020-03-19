package core

import (
	"sync"
	"sync/atomic"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

type DeltaChangeEntry struct {
	isSell bool
	price  sdk.Dec
	amount sdk.Int
}

// Each market needs three managers
type TripleManager struct {
	sell *DepthManager
	buy  *DepthManager
	tkm  *TickerManager

	// Depth data are updated in a batch mode, i.e., one block applies as a whole
	// So these data can not be queried during the execution of a block
	// We use this mutex to protect them from being queried if they are during updating of a block
	mutex sync.RWMutex

	entryList []DeltaChangeEntry
}

func (triman *TripleManager) AddDeltaChange(isSell bool, price sdk.Dec, amount sdk.Int) {
	triman.entryList = append(triman.entryList, DeltaChangeEntry{isSell, price, amount})
}

func (triman *TripleManager) Update(lockCount *int64) {
	if len(triman.entryList) == 0 {
		return
	}
	triman.mutex.Lock()
	defer triman.mutex.Unlock()
	atomic.AddInt64(lockCount, 1)
	for _, e := range triman.entryList {
		if e.isSell {
			triman.sell.DeltaChange(e.price, e.amount)
		} else {
			triman.buy.DeltaChange(e.price, e.amount)
		}
	}
	triman.entryList = triman.entryList[:0]
}
