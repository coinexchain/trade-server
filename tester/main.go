// +build tester

package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/coinexchain/trade-server/core"
	sdk "github.com/cosmos/cosmos-sdk/types"
	dbm "github.com/tendermint/tm-db"
)

type dealRec struct {
	price  sdk.Dec
	amount int64
	time   int64
}

func getRandDealRecList(size int, seed int64, timeStep int32) []dealRec {
	r := rand.New(rand.NewSource(seed))
	currTime := T("2019-08-21T00:00:00.0Z").Unix()
	recList := make([]dealRec, size)
	for i := 0; i < size; i++ {
		recList[i].price = sdk.NewDec(r.Int63())
		recList[i].amount = r.Int63()
		currTime += int64(r.Int31n(timeStep))
		recList[i].time = currTime
	}
	return recList
}

func getCandleStick(recList []dealRec, span string) core.CandleStick {
	low := recList[0].price
	high := recList[0].price
	totalDeal := sdk.NewInt(recList[0].amount)
	for i := 1; i < len(recList); i++ {
		if recList[i].price.GT(high) {
			high = recList[i].price
		}
		if recList[i].price.LT(low) {
			low = recList[i].price
		}
		totalDeal = totalDeal.AddRaw(recList[i].amount)
	}
	return core.CandleStick{
		OpenPrice:      recList[0].price,
		ClosePrice:     recList[len(recList)-1].price,
		HighPrice:      high,
		LowPrice:       low,
		TotalDeal:      totalDeal,
		EndingUnixTime: recList[len(recList)-1].time,
		TimeSpan:       span,
		Market:         "",
	}
}

type CandleStickMan struct {
	MinuteList []core.CandleStick
	HourList   []core.CandleStick
	DayList    []core.CandleStick
}

func (csMan *CandleStickMan) scanForMinutes(recList []dealRec) {
	lastIdx := 0
	lastTime := time.Unix(recList[0].time, 0)
	var currIdx int
	for currIdx = 1; currIdx < len(recList); currIdx++ {
		currTime := time.Unix(recList[currIdx].time, 0)
		if currTime.UTC().Minute() != lastTime.UTC().Minute() {
			cs := getCandleStick(recList[lastIdx:currIdx], core.MinuteStr)
			csMan.MinuteList = append(csMan.MinuteList, cs)
			lastIdx = currIdx
			lastTime = currTime
		}
	}
	if lastIdx != currIdx {
		cs := getCandleStick(recList[lastIdx:currIdx], core.MinuteStr)
		csMan.MinuteList = append(csMan.MinuteList, cs)
	}
}

func (csMan *CandleStickMan) scanForHours(recList []dealRec) {
	lastIdx := 0
	lastTime := time.Unix(recList[0].time, 0)
	var currIdx int
	for currIdx = 1; currIdx < len(recList); currIdx++ {
		currTime := time.Unix(recList[currIdx].time, 0)
		if currTime.UTC().Hour() != lastTime.UTC().Hour() {
			cs := getCandleStick(recList[lastIdx:currIdx], core.HourStr)
			csMan.HourList = append(csMan.HourList, cs)
			lastIdx = currIdx
			lastTime = currTime
		}
	}
	if lastIdx != currIdx {
		cs := getCandleStick(recList[lastIdx:currIdx], core.HourStr)
		csMan.HourList = append(csMan.HourList, cs)
	}
}

func (csMan *CandleStickMan) scanForDays(recList []dealRec) {
	lastIdx := 0
	lastTime := time.Unix(recList[0].time, 0)
	var currIdx int
	for currIdx = 1; currIdx < len(recList); currIdx++ {
		currTime := time.Unix(recList[currIdx].time, 0)
		if currTime.UTC().Day() != lastTime.UTC().Day() {
			cs := getCandleStick(recList[lastIdx:currIdx], core.DayStr)
			csMan.DayList = append(csMan.DayList, cs)
			lastIdx = currIdx
			lastTime = currTime
		}
	}
	if lastIdx != currIdx {
		cs := getCandleStick(recList[lastIdx:currIdx], core.DayStr)
		csMan.DayList = append(csMan.DayList, cs)
	}
}

func testCandleStick(recList []dealRec) {
	fmt.Printf("Begin testCandleStick\n")
	//for i, rec := range recList {
	//	t := time.Unix(rec.time, 0)
	//	fmt.Printf("H %d p %s a %d t %d %d-%d:%d:%d\n", i, rec.price.String(), rec.amount, rec.time,
	//		t.Day(), t.Hour(), t.Minute(), t.Second())
	//}
	refMan := CandleStickMan{
		MinuteList: make([]core.CandleStick, 0, 10000),
		HourList:   make([]core.CandleStick, 0, 200),
		DayList:    make([]core.CandleStick, 0, 10),
	}
	refMan.scanForMinutes(recList)
	refMan.scanForHours(recList)
	refMan.scanForDays(recList)

	impMan := core.NewCandleStickManager([]string{""})
	impMList := make([]core.CandleStick, 0, 10000)
	impHList := make([]core.CandleStick, 0, 200)
	impDList := make([]core.CandleStick, 0, 10)
	var t time.Time
	for _, deal := range recList {
		t = time.Unix(deal.time, 0)
		csList := impMan.NewBlock(t)
		for _, cs := range csList {
			if cs.TotalDeal.IsZero() {
				continue
			}
			if cs.TimeSpan == core.MinuteStr {
				impMList = append(impMList, *cs)
			}
			if cs.TimeSpan == core.HourStr {
				impHList = append(impHList, *cs)
			}
			if cs.TimeSpan == core.DayStr {
				impDList = append(impDList, *cs)
			}
		}
		impMan.GetRecord("").Update(t, deal.price, deal.amount)
	}
	//flush
	t = time.Unix(t.Unix()+60*60*24, 0)
	csList := impMan.NewBlock(t)
	for _, cs := range csList {
		if cs.TimeSpan == core.MinuteStr {
			impMList = append(impMList, *cs)
		}
		if cs.TimeSpan == core.HourStr {
			impHList = append(impHList, *cs)
		}
		if cs.TimeSpan == core.DayStr {
			impDList = append(impDList, *cs)
		}
	}
	compareCandleSticks(refMan.MinuteList, impMList, "minute")
	compareCandleSticks(refMan.HourList, impHList, "hour")
	compareCandleSticks(refMan.DayList, impDList, "day")
}

func compareCandleSticks(ref, imp []core.CandleStick, span string) {
	for i := 0; i < len(ref) && i < len(imp); i++ {
		if ref[i].OpenPrice != imp[i].OpenPrice {
			panic("OpenPrice mismatch")
		}
		if ref[i].ClosePrice != imp[i].ClosePrice {
			fmt.Printf("S %s E: %d  %s %s\n", span, ref[i].EndingUnixTime, ref[i].ClosePrice, imp[i].ClosePrice)
			panic("ClosePrice mismatch")
		}
		if ref[i].HighPrice != imp[i].HighPrice {
			panic("HighPrice mismatch")
		}
		if ref[i].LowPrice != imp[i].LowPrice {
			panic("LowPrice mismatch")
		}
		if ref[i].EndingUnixTime != imp[i].EndingUnixTime {
			panic("EndingUnixTime mismatch")
		}
		if !ref[i].TotalDeal.Equal(imp[i].TotalDeal) {
			fmt.Printf("E: %d  %s %s\n", ref[i].EndingUnixTime, ref[i].TotalDeal, imp[i].TotalDeal)
			panic("TotalDeal mismatch")
		}
	}
	if len(ref) != len(imp) {
		for i := 0; i < len(ref); i++ {
			fmt.Printf("X %d %d\n", i, ref[i].EndingUnixTime)
		}
		for i := 0; i < len(imp); i++ {
			fmt.Printf("Y %d %d\n", i, imp[i].EndingUnixTime)
		}
		panic("length mismatch")
	}
}

//==========================================

type DepthManager struct {
	ppMap map[string]core.PricePoint
}

func getRandPricePointsSets(count int, sizeLimit int32, seed int64, priceRange int32, amountRange int32) [][]core.PricePoint {
	r := rand.New(rand.NewSource(seed))
	res := make([][]core.PricePoint, count)
	for i := 0; i < count; i++ {
		res[i] = getRandPricePoints(int(r.Int31n(sizeLimit)), r, priceRange, amountRange)
	}
	return res
}

func getRandPricePoints(size int, r *rand.Rand, priceRange int32, amountRange int32) []core.PricePoint {
	ppList := make([]core.PricePoint, size)
	for i := 0; i < size; i++ {
		a := int64(r.Int31n(amountRange))
		if r.Int31n(2)%2 == 0 {
			ppList[i].Amount = sdk.NewInt(a)
		} else {
			ppList[i].Amount = sdk.NewInt(-a)
		}
		p := int64(r.Int31n(priceRange))
		ppList[i].Price = sdk.NewDec(p)
	}
	return ppList
}

func (dm *DepthManager) Update(points []core.PricePoint) {
	for _, pp := range points {
		s := pp.Price.String()
		oldPP, ok := dm.ppMap[s]
		if ok {
			oldPP.Amount = oldPP.Amount.Add(pp.Amount)
			dm.ppMap[s] = oldPP
		} else {
			dm.ppMap[s] = pp
		}
		if dm.ppMap[s].Amount.IsZero() {
			delete(dm.ppMap, s)
		}
	}
}

func (dm *DepthManager) GetSortedPoints() []core.PricePoint {
	points := make([]core.PricePoint, 0, len(dm.ppMap))
	for _, pp := range dm.ppMap {
		points = append(points, pp)
	}
	sort.Slice(points, func(i, j int) bool {
		return points[i].Price.LT(points[j].Price)
	})
	return points
}

func testDepth(pointsSets [][]core.PricePoint) {
	fmt.Printf("Begin testDepth\n")
	//for i, points := range pointsSets {
	//	for j, point := range points {
	//		fmt.Printf("H %d %d %s %s\n", i, j, point.Price.String(), point.Amount.String())
	//	}
	//}
	refMan := DepthManager{ppMap: make(map[string]core.PricePoint)}
	impMan := core.NewDepthManager("")
	for x, points := range pointsSets {
		changes := make(map[string]core.PricePoint)
		for _, point := range points {
			changes[point.Price.String()] = point
		}
		refMan.Update(points)

		for _, point := range points {
			impMan.DeltaChange(point.Price, point.Amount)
		}
		ppMap, _ := impMan.EndBlock()

		cL := make([]string, 0, len(changes))
		for s := range changes {
			cL = append(cL, s)
		}

		pL := make([]string, 0, len(ppMap))
		for _, pp := range ppMap {
			pL = append(pL, pp.Price.String())
			_, ok := changes[pp.Price.String()]
			if !ok {
				panic("key mismatch")
			}
		}

		sort.Strings(cL)
		sort.Strings(pL)

		refPricePoints := refMan.GetSortedPoints()
		impPricePoints := impMan.GetLowest(impMan.Size())
		if len(changes) != len(ppMap) {
			for i := 0; i < len(refPricePoints); i++ {
				fmt.Printf("ref %d %d %s %s\n", x, i, refPricePoints[i].Price, refPricePoints[i].Amount)
			}
			for i := 0; i < len(impPricePoints); i++ {
				fmt.Printf("imp %d %d %s %s\n", x, i, impPricePoints[i].Price, impPricePoints[i].Amount)
			}
			fmt.Printf("length mismatch %d %d\n", len(changes), len(ppMap))
			for _, s := range cL {
				fmt.Printf("A: %s\n", s)
			}
			for _, s := range pL {
				fmt.Printf("B: %s\n", s)
			}
			panic("length mismatch")
		}

		for i := 0; i < len(refPricePoints) && i < len(impPricePoints); i++ {
			//fmt.Printf("h %d %d %s %s VS %s %s\n", x, i, refPricePoints[i].Price,refPricePoints[i].Amount,
			//        impPricePoints[i].Price,impPricePoints[i].Amount)
			if !refPricePoints[i].Price.Equal(impPricePoints[i].Price) {
				panic("Price mismatch")
			}
			if !refPricePoints[i].Amount.Equal(impPricePoints[i].Amount) {
				panic("Amount mismatch")
			}
		}
		if len(refPricePoints) != len(impPricePoints) {
			panic("length mismatch")
		}
	}
}

//==========================================

func getRandPriceList(size int, seed int64, step int32, priceRange int32) []sdk.Dec {
	r := rand.New(rand.NewSource(seed))
	priceList := make([]sdk.Dec, size)
	var stripe int
	for i := 0; i < size; i += stripe {
		p := sdk.NewDec(int64(r.Int31n(priceRange)))
		stripe = int(r.Int31n(step))
		if i+stripe >= size {
			stripe = size - i
		}
		for j := 0; j < stripe; j++ {
			priceList[i+j] = p
		}
	}
	return priceList
}

func findTickers(priceList []sdk.Dec) ([]core.Ticker, []int) {
	resPos := make([]int, 0, 1000)
	res := make([]core.Ticker, 0, 1000)
	if len(priceList) <= core.MinuteNumInDay {
		panic("priceList too small!")
	}
	//resPos[0] = core.MinuteNumInDay
	//res[0] = core.Ticker{
	//	NewPrice:          priceList[core.MinuteNumInDay],
	//	OldPriceOneDayAgo: priceList[0],
	//	Market:            "",
	//}
	for i := core.MinuteNumInDay + 1; i < len(priceList); i++ {
		j := i - core.MinuteNumInDay
		if priceList[j].Equal(priceList[j-1]) && priceList[i].Equal(priceList[i-1]) {
			continue
		}
		resPos = append(resPos, i)
		res = append(res, core.Ticker{
			NewPrice:          priceList[i],
			OldPriceOneDayAgo: priceList[j],
			Market:            "",
		})
	}
	return res, resPos
}

func testTicker(priceList []sdk.Dec) {
	fmt.Printf("Begin testTicker\n")
	refTickers, refIdxList := findTickers(priceList)
	tkMan := core.NewTickerManager("")
	for i := 0; i < core.MinuteNumInDay; i++ {
		tkMan.UpdateNewestPrice(priceList[i], i)
	}
	impTickers := make([]core.Ticker, 0, 1000)
	impIdxList := make([]int, 0, 1000)
	for i := core.MinuteNumInDay; i < len(priceList); i++ {
		j := i % core.MinuteNumInDay
		tkMan.UpdateNewestPrice(priceList[i], j)
		ticker := tkMan.GetTicker(j)
		if ticker != nil {
			impTickers = append(impTickers, *ticker)
			impIdxList = append(impIdxList, i)
		}
	}
	for i := 0; i < len(refTickers) && i < len(impTickers); i++ {
		if !refTickers[i].NewPrice.Equal(impTickers[i].NewPrice) {
			fmt.Printf("Ref %d: %d %d %s %s\n", i, refIdxList[i], refIdxList[i]-core.MinuteNumInDay, refTickers[i].NewPrice, refTickers[i].OldPriceOneDayAgo)
			fmt.Printf("Imp %d: %d %d %s %s\n", i, impIdxList[i], impIdxList[i]-core.MinuteNumInDay, impTickers[i].NewPrice, impTickers[i].OldPriceOneDayAgo)
			panic("NewPrice not equal")
		}
		if !refTickers[i].OldPriceOneDayAgo.Equal(impTickers[i].OldPriceOneDayAgo) {
			panic("OldPriceOneDayAgo not equal")
		}
	}
	for i := 0; i < len(refIdxList) && i < len(impIdxList); i++ {
		if refIdxList[i] != impIdxList[i] {
			panic("Idx not equal")
		}
	}
	if len(refTickers) != len(impTickers) {
		panic("length not equal")
	}
	if len(refIdxList) != len(impIdxList) {
		panic("length not equal")
	}
}

//==========================================

func getRandPriceAndAmountList(size int, seed int64, step int32, priceRange int32, amountRange int32) ([]sdk.Dec, []sdk.Int) {
	r := rand.New(rand.NewSource(seed))
	priceList := make([]sdk.Dec, size)
	amountList := make([]sdk.Int, size)
	var stripe int
	for i := 0; i < size; i += stripe {
		p := sdk.NewDec(int64(r.Int31n(priceRange)) + 1)
		stripe = int(r.Int31n(step))
		if i+stripe >= size {
			stripe = size - i
		}
		for j := 0; j < stripe; j++ {
			priceList[i+j] = p
		}
		a := sdk.NewInt(int64(r.Int31n(amountRange)))
		amountList[i] = a
		for j := 1; j < stripe; j++ {
			amountList[i+j] = sdk.ZeroInt()
		}
	}
	return priceList, amountList
}

func findXTickers(priceList []sdk.Dec, amountList []sdk.Int) ([]core.XTicker, []int) {
	resPos := make([]int, 0, 1000)
	res := make([]core.XTicker, 0, 1000)
	if len(priceList) <= core.MinuteNumInDay {
		panic("priceList too small!")
	}
	for i := core.MinuteNumInDay; i < len(priceList); i++ {
		j := i - core.MinuteNumInDay
		totalAmount := sdk.ZeroInt()
		highPrice := sdk.ZeroDec()
		for k := j + 1; k <= i; k++ {
			totalAmount = totalAmount.Add(amountList[k])
			//if i == 1440 && !amountList[k].IsZero() {
			//	fmt.Printf("Amount: %d %s %s\n", k, totalAmount, amountList[k])
			//}
			//if i == 1933 {
			//	fmt.Printf("HighPrice: %d %s %s\n", k, priceList[k], highPrice)
			//}
			if !amountList[k].IsZero() && priceList[k].GT(highPrice) {
				highPrice = priceList[k]
			}
		}
		lowPrice := highPrice
		for k := j + 1; k <= i; k++ {
			//if i == 8505 {
			//	fmt.Printf("LowPrice: %d %s %s\n", k, priceList[k], lowPrice)
			//}
			if !amountList[k].IsZero() && priceList[k].LT(lowPrice) {
				lowPrice = priceList[k]
			}
		}
		resPos = append(resPos, i)
		res = append(res, core.XTicker{
			NewPrice:          priceList[i],
			OldPriceOneDayAgo: priceList[j],
			HighPrice:         highPrice,
			LowPrice:          lowPrice,
			TotalDeal:         totalAmount,
			Market:            "",
		})
	}
	return res, resPos
}

func testXTicker(priceList []sdk.Dec, amountList []sdk.Int) {
	fmt.Printf("Begin testXTicker\n")
	refXTickers, refIdxList := findXTickers(priceList, amountList)
	tkMan := core.NewXTickerManager("")
	for i := 0; i < core.MinuteNumInDay; i++ {
		tkMan.UpdateNewestPrice(priceList[i], i, amountList[i])
	}
	impXTickers := make([]core.XTicker, 0, 1000)
	impIdxList := make([]int, 0, 1000)
	//target := 8505
	//for i := target-core.MinuteNumInDay; i <= target; i++ {
	//	fmt.Printf("%% %d %d p: %s a: %s\n", i, i%core.MinuteNumInDay, priceList[i], amountList[i])
	//}
	for i := core.MinuteNumInDay; i < len(priceList); i++ {
		j := i % core.MinuteNumInDay
		tkMan.UpdateNewestPrice(priceList[i], j, amountList[i])
		ticker := tkMan.GetXTicker(j)
		if ticker != nil {
			impXTickers = append(impXTickers, *ticker)
			impIdxList = append(impIdxList, i)
		}
	}
	for i := 0; i < len(refXTickers) && i < len(impXTickers); i++ {
		if !refXTickers[i].NewPrice.Equal(impXTickers[i].NewPrice) {
			panic("NewPrice not equal")
		}
		if !refXTickers[i].OldPriceOneDayAgo.Equal(impXTickers[i].OldPriceOneDayAgo) {
			panic("OldPriceOneDayAgo not equal")
		}
		if !refXTickers[i].TotalDeal.Equal(impXTickers[i].TotalDeal) {
			fmt.Printf("Ref %d: %d %d %s %s\n", i, refIdxList[i], refIdxList[i]-core.MinuteNumInDay, refXTickers[i].NewPrice, refXTickers[i].TotalDeal)
			fmt.Printf("Imp %d: %d %d %s %s\n", i, impIdxList[i], impIdxList[i]-core.MinuteNumInDay, impXTickers[i].NewPrice, impXTickers[i].TotalDeal)
			panic("TotalDeal not equal")
		}
		if !refXTickers[i].TotalDeal.IsZero() && !refXTickers[i].HighPrice.Equal(impXTickers[i].HighPrice) {
			fmt.Printf("Ref %d: %d %d %s %s\n", i, refIdxList[i], refIdxList[i]-core.MinuteNumInDay, refXTickers[i].NewPrice, refXTickers[i].HighPrice)
			fmt.Printf("Imp %d: %d %d %s %s\n", i, impIdxList[i], impIdxList[i]-core.MinuteNumInDay, impXTickers[i].NewPrice, impXTickers[i].HighPrice)
			panic("HighPrice not equal")
		}
		if !refXTickers[i].TotalDeal.IsZero() && !refXTickers[i].LowPrice.Equal(impXTickers[i].LowPrice) {
			fmt.Printf("Ref %d: %d %d %s %s\n", i, refIdxList[i], refIdxList[i]-core.MinuteNumInDay, refXTickers[i].NewPrice, refXTickers[i].LowPrice)
			fmt.Printf("Imp %d: %d %d %s %s\n", i, impIdxList[i], impIdxList[i]-core.MinuteNumInDay, impXTickers[i].NewPrice, impXTickers[i].LowPrice)
			panic("LowPrice not equal")
		}
	}
	for i := 0; i < len(refIdxList) && i < len(impIdxList); i++ {
		if refIdxList[i] != impIdxList[i] {
			panic("Idx not equal")
		}
	}
	if len(refXTickers) != len(impXTickers) {
		panic("length not equal")
	}
	if len(refIdxList) != len(impIdxList) {
		panic("length not equal")
	}
}

//----------------------------------------

func toStr(payload []json.RawMessage) string {
	out := make([]string, len(payload))
	for i := 0; i < len(out); i++ {
		out[i] = string(payload[i])
	}
	return strings.Join(out, "\n")
}

func T(s string) time.Time {
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		panic(err)
	}
	return t
}

func simulateKafkaInput() {
	// defer profile.Start().Stop()
	println("Args[0]: ", os.Args[0])
	file, err := os.Open(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	db := dbm.NewMemDB()
	subMan := core.GetSubscribeManager("coinex1x6rhu5m53fw8qgpwuljauaptvxyur57zym4jly", "coinex1yj66ancalgk7dz3383s6cyvdd0nd93q0tk4x0c")
	//hub := core.NewHub(db, subMan, 60, -1, -1, 4447714, "", 4545600)
	hub := core.NewHub(db, subMan, 60, -1, -1, 0, "", 99999999)

	scanner := bufio.NewScanner(file)
	size := 100 * 1024 * 1024
	scanner.Buffer(make([]byte, size), size)
	counter := int64(0)
	for scanner.Scan() {
		counter++
		if counter%10000 == 0 {
			println("==========", counter)
		}
		//if counter > 400000 {
		//	break
		//}
		line := scanner.Text()
		divIdx := strings.Index(line, "#")
		msgType := line[:divIdx]
		msg := line[divIdx+1:]
		//fmt.Printf("%s %s\n", msgType, msg)
		hub.ConsumeMessage(msgType, []byte(msg))
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	//unixTime := T("2019-09-29T08:02:06.647266Z").Unix()
	//data := hub.QueryCandleStick("hffp/cet", core.Hour, unixTime, 0, 1000)
	//fmt.Printf("here %s %d\n", toStr(data), unixTime)

	//unixTime = T("2019-09-29T08:02:06.647266Z").Unix()
	//data = hub.QueryCandleStick("hffp/cet", core.Day, unixTime, 0, 1000)
	//fmt.Printf("here %s %d\n", toStr(data), unixTime)

	info := hub.QueryTxByHashID("E26ACB4D854DE5609FF3AA16F44B2EF9A46757B069C14723DBD1A2A0903012B3")
	fmt.Printf("=hhhh= %s\n", info)

	//t := T("2021-09-29T08:02:06.647266Z").Unix()
	//acc := "coinex1m77xwm4t29ga2xg2dzrwzm6wf2mumw3gzl4yd6"
	//msgList, timesid := hub.QueryIncomeAboutToken("cet", acc, t, 0, 1024)
	//for i, msg := range msgList {
	//	fmt.Printf("== %s %d\n", string(msg), timesid[2*i])
	//}
	//fmt.Printf("------------------------\n")
	//acc = "coinex1m0h3aw8n83tmlyltqldj7xqyp00wdvldxlw3nu"
	//msgList, timesid = hub.QueryTxAboutToken("cet", acc, t, 0, 1024)
	//for i, msg := range msgList {
	//	fmt.Printf("== %s %d\n", string(msg), timesid[2*i])
	//}

	//t := T("2020-09-29T08:02:06.647266Z").Unix()
	//msgList, timesid := hub.QueryDeal("abc/cet", t, 0, 1024)
	//for i, msg := range msgList {
	//	fmt.Printf("== %s %d\n", string(msg), timesid[2*i])
	//}
	//msgList = hub.QueryCandleStick("abc/cet", core.Minute, t, 0, 1024)
	//for _, msg := range msgList {
	//	fmt.Printf("== %s\n", string(msg))
	//}
	//msgList = hub.QueryCandleStick("abc/cet", core.Hour, t, 0, 1024)
	//for _, msg := range msgList {
	//	fmt.Printf("== %s\n", string(msg))
	//}
	//msgList = hub.QueryCandleStick("abc/cet", core.Day, t, 0, 1024)
	//for _, msg := range msgList {
	//	fmt.Printf("== %s\n", string(msg))
	//}

	//tickers := hub.QueryTickers([]string{"abc/cet", "hffp/cet", "game/cet", "hffp1/cet", "ddc/cet", "ift/cet", "ktt/cet"})
	//for _, ticker := range tickers {
	//	fmt.Printf("== %v\n", ticker)
	//}
}

func main() {
	if len(os.Args) != 2 && len(os.Args) != 1 {
		fmt.Fprintf(os.Stderr, "usage: %s [inputfile]\n", os.Args[0])
		os.Exit(2)
	}
	if len(os.Args) == 2 {
		simulateKafkaInput()
		return
	}
	fmt.Printf("Now run random test\n")
	//                                  size    seed    timeStep
	testCandleStick(getRandDealRecList(50000, 0, 40))
	testCandleStick(getRandDealRecList(90000, 1, 20))
	testCandleStick(getRandDealRecList(90000, 2, 90))
	testCandleStick(getRandDealRecList(90000, 3, 50))
	testCandleStick(getRandDealRecList(90000, 4, 150))
	testCandleStick(getRandDealRecList(90000, 15, 10))
	testCandleStick(getRandDealRecList(990000, 25, 30))
	testCandleStick(getRandDealRecList(990000, 27, 20))
	testCandleStick(getRandDealRecList(990000, 28, 25))
	testCandleStick(getRandDealRecList(990000, 29, 27))

	//                                                         seed  step priceRange amountRange
	pList, aList := getRandPriceAndAmountList(core.MinuteNumInDay*7, 0, 50, 1000, 100)
	testXTicker(pList, aList)
	pList, aList = getRandPriceAndAmountList(core.MinuteNumInDay*7, 5, 16, 200, 100)
	testXTicker(pList, aList)
	pList, aList = getRandPriceAndAmountList(core.MinuteNumInDay*7, 10, 150, 400, 100)
	testXTicker(pList, aList)
	pList, aList = getRandPriceAndAmountList(core.MinuteNumInDay*7, 11, 30, 400, 100)
	testXTicker(pList, aList)
	pList, aList = getRandPriceAndAmountList(core.MinuteNumInDay*7, 12, 20, 300, 100)
	testXTicker(pList, aList)
	pList, aList = getRandPriceAndAmountList(core.MinuteNumInDay*7, 13, 120, 300, 100)
	testXTicker(pList, aList)
	pList, aList = getRandPriceAndAmountList(core.MinuteNumInDay*7, 14, 920, 300, 100)
	testXTicker(pList, aList)
	pList, aList = getRandPriceAndAmountList(core.MinuteNumInDay*7, 15, 920, 300, 100)
	testXTicker(pList, aList)
	pList, aList = getRandPriceAndAmountList(core.MinuteNumInDay*177, 16, 20, 50, 100)
	testXTicker(pList, aList)
	pList, aList = getRandPriceAndAmountList(core.MinuteNumInDay*177, 17, 20, 50, 100)
	testXTicker(pList, aList)

	//                                                  seed  step priceRange
	testTicker(getRandPriceList(core.MinuteNumInDay*7, 0, 50, 1000))
	testTicker(getRandPriceList(core.MinuteNumInDay*7, 5, 16, 200))
	testTicker(getRandPriceList(core.MinuteNumInDay*7, 10, 150, 400))
	testTicker(getRandPriceList(core.MinuteNumInDay*7, 11, 30, 400))
	testTicker(getRandPriceList(core.MinuteNumInDay*7, 12, 20, 300))
	testTicker(getRandPriceList(core.MinuteNumInDay*7, 13, 120, 300))
	testTicker(getRandPriceList(core.MinuteNumInDay*7, 14, 920, 300))
	testTicker(getRandPriceList(core.MinuteNumInDay*7, 15, 920, 300))
	testTicker(getRandPriceList(core.MinuteNumInDay*177, 16, 20, 50))
	testTicker(getRandPriceList(core.MinuteNumInDay*177, 17, 20, 50))

	//                               count   sizeLimit seed  priceRange amountRange
	testDepth(getRandPricePointsSets(100, 20, 0, 100, 20))
	testDepth(getRandPricePointsSets(1100, 40, 11, 200, 50))
	testDepth(getRandPricePointsSets(1200, 30, 12, 200, 50))
	testDepth(getRandPricePointsSets(1200, 50, 22, 100, 150))
	testDepth(getRandPricePointsSets(1200, 50, 23, 100, 120))
	testDepth(getRandPricePointsSets(91200, 10, 24, 100, 120))
	testDepth(getRandPricePointsSets(99200, 10, 25, 1100, 120))
	testDepth(getRandPricePointsSets(99200, 10, 26, 1100, 120))
	testDepth(getRandPricePointsSets(99200, 10, 27, 1100, 100))
	testDepth(getRandPricePointsSets(99200, 10, 28, 1100, 20))

}
