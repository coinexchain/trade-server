package core

import (
	"math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/emirpasic/gods/maps/treemap"
)

// Manager for the depth information of one side of the order book: sell or buy
type DepthManager struct {
	ppMap      *treemap.Map //map[string]*PricePoint
	Updated    map[string]*PricePoint
	Side       string
	Levels     []string
	MulDecs    []sdk.Dec
	LevelDepth map[string]map[string]*PricePoint
}

func NewDepthManager(side string) *DepthManager {
	dm := &DepthManager{
		ppMap:      treemap.NewWithStringComparator(),
		Updated:    make(map[string]*PricePoint),
		Side:       side,
		Levels:     make([]string, 0, 20),
		MulDecs:    make([]sdk.Dec, 0, 20),
		LevelDepth: make(map[string]map[string]*PricePoint, 20),
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
	p, err := sdk.NewDecFromStr(levelStr)
	if err != nil {
		return err
	}
	dm.MulDecs = append(dm.MulDecs, sdk.OneDec().QuoTruncate(p))
	dm.Levels = append(dm.Levels, levelStr)
	dm.LevelDepth[levelStr] = make(map[string]*PricePoint)
	return nil
}

func (dm *DepthManager) Size() int {
	return dm.ppMap.Size()
}

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

// positive amount for increment, negative amount for decrement
func (dm *DepthManager) DeltaChange(price sdk.Dec, amount sdk.Int) {
	s := string(DecToBigEndianBytes(price))
	ptr, ok := dm.ppMap.Get(s)
	var pp *PricePoint
	if !ok {
		pp = &PricePoint{Price: price, Amount: sdk.ZeroInt()}
	} else {
		pp = ptr.(*PricePoint)
	}
	// tmp := pp.Amount
	pp.Amount = pp.Amount.Add(amount)
	if pp.Amount.IsZero() {
		dm.ppMap.Remove(s)
	} else {
		dm.ppMap.Put(s, pp)
	}
	//if "8.800000000000000000" == price.String() && "buy" == dm.Side {
	// fmt.Printf("== %s Price: %s amount: %s %s => %s\n", dm.Side, price.String(), amount.String(), tmp.String(), pp.Amount.String())
	//}
	dm.Updated[s] = pp

	isBuy := true
	if dm.Side == "sell" {
		isBuy = false
	}
	for i, lev := range dm.Levels {
		updateAmount(dm.LevelDepth[lev], &PricePoint{Price: price, Amount: amount}, dm.MulDecs[i], isBuy)
	}
}

// returns the changed PricePoints of last block. Clear dm.Updated for the next block
func (dm *DepthManager) EndBlock() (map[string]*PricePoint, map[string]map[string]*PricePoint) {
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
