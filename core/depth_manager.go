package core

import (
	"math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/emirpasic/gods/maps/treemap"
)

// Manager for the depth information of one side of the order book: sell or buy
type DepthManager struct {
	// ppMap keeps a depth map without any merging
	ppMap   *treemap.Map           // we can not use map[string]*PricePoint because iteration is needed
	Updated map[string]*PricePoint //it records all the changed PricePoint during last block
	Side    string
	//following members are used for depth merging
	Levels                 []string
	CachedMulDecsForLevels []sdk.Dec
	LevelDepth             LevelsPricePoint
}

func NewDepthManager(side string) *DepthManager {
	dm := &DepthManager{
		ppMap:                  treemap.NewWithStringComparator(),
		Updated:                make(map[string]*PricePoint),
		Side:                   side,
		Levels:                 make([]string, 0, 20),
		CachedMulDecsForLevels: make([]sdk.Dec, 0, 20),
		LevelDepth:             make(LevelsPricePoint, 20),
	}
	return dm
}

func (dm *DepthManager) AddLevel(levelStr string) error {
	for _, lvl := range dm.Levels {
		//Return if already added
		if lvl == levelStr {
			return nil
		}
	}
	if levelStr == "all" {
		dm.CachedMulDecsForLevels = append(dm.CachedMulDecsForLevels, sdk.OneDec())
	} else {
		p, err := sdk.NewDecFromStr(levelStr)
		if err != nil {
			return err
		}
		dm.CachedMulDecsForLevels = append(dm.CachedMulDecsForLevels, sdk.OneDec().QuoTruncate(p))
	}
	dm.Levels = append(dm.Levels, levelStr)
	dm.LevelDepth[levelStr] = make(map[string]*PricePoint)
	return nil
}

func (dm *DepthManager) Size() int {
	return dm.ppMap.Size()
}

// dump the PricePoints in ppMap for serialization
func (dm *DepthManager) DumpPricePoints() []*PricePoint {
	if dm == nil {
		return nil
	}
	size := dm.ppMap.Size()
	pps := make([]*PricePoint, size)
	iter := dm.ppMap.Iterator()
	iter.Begin()
	for i := 0; i < size; i++ {
		iter.Next()
		pps[i] = iter.Value().(*PricePoint)
	}
	return pps
}

// The orderbook's entry at 'price' is delta-changed by amount
// positive amount for increment, negative amount for decrement
func (dm *DepthManager) DeltaChange(price sdk.Dec, amount sdk.Int) {
	s := string(decToBigEndianBytes(price))
	ptr, ok := dm.ppMap.Get(s)
	var pp *PricePoint
	if !ok {
		pp = &PricePoint{Price: price, Amount: sdk.ZeroInt()}
	} else {
		pp = ptr.(*PricePoint)
	}
	pp.Amount = pp.Amount.Add(amount)
	if pp.Amount.IsZero() {
		dm.ppMap.Remove(s)
	} else {
		dm.ppMap.Put(s, pp)
	}
	// update the depth map without merging
	dm.Updated[s] = pp

	isBuy := true
	if dm.Side == "sell" {
		isBuy = false
	}
	for i, lev := range dm.Levels {
		// update the depth maps with merging
		updateAmount(dm.LevelDepth[lev],
			&PricePoint{Price: price, Amount: amount},
			dm.CachedMulDecsForLevels[i],
			isBuy,
		)
	}
}

// Returns the changed PricePoints of last block, without and with merging
// Clear dm.Updated and dm.LevelDepth for the next block
func (dm *DepthManager) EndBlock() (woMerging map[string]*PricePoint, wiMerging map[string]map[string]*PricePoint) {
	oldUpdated, oldLevelDepth := dm.Updated, dm.LevelDepth
	dm.Updated = make(map[string]*PricePoint)
	dm.LevelDepth = make(LevelsPricePoint, len(dm.Levels))
	for _, lev := range dm.Levels {
		dm.LevelDepth[lev] = make(map[string]*PricePoint)
	}
	return oldUpdated, oldLevelDepth
}

// Returns the highest n PricePoints
func (dm *DepthManager) GetHighest(n int) []*PricePoint {
	res := make([]*PricePoint, 0, n)
	iter := dm.ppMap.Iterator()
	iter.End()
	if n == 0 {
		n = math.MaxInt64
	}
	for i := 0; i < n; i++ {
		if ok := iter.Prev(); !ok {
			break
		}
		res = append(res, iter.Value().(*PricePoint))
	}
	return res
}

// Returns the lowest n PricePoints
func (dm *DepthManager) GetLowest(n int) []*PricePoint {
	res := make([]*PricePoint, 0, n)
	iter := dm.ppMap.Iterator()
	iter.Begin()
	if n == 0 {
		n = math.MaxInt64
	}
	for i := 0; i < n; i++ {
		if ok := iter.Next(); !ok {
			break
		}
		res = append(res, iter.Value().(*PricePoint))
	}
	return res
}
