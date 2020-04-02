package core

// We offload the logic of pushing subscribers to this goroutine,
// such that ConsumeMessage can run a little faster
func (hub *Hub) pushMsgToWebsocket() {
	for {
		entry := <-hub.msgsChannel
		switch entry.topic {
		case BlockInfoKey:
			hub.PushHeightInfoMsg(entry.bz)
		case KlineKey:
			vals := entry.extra.([]string)
			hub.PushCandleMsg( /*market*/ vals[0], entry.bz /*timespan*/, vals[1])
		case LockedKey:
			hub.PushLockedCoinsMsg( /*addr*/ entry.extra.(string), entry.bz)
		case IncomeKey:
			hub.PushIncomeMsg( /*receiver*/ entry.extra.(string), entry.bz)
		case TxKey:
			hub.PushTxMsg( /*addr*/ entry.extra.(string), entry.bz)
		case RedelegationKey:
			hub.PushRedelegationMsg(entry.extra.(TimeAndSidWithAddr))
		case UnbondingKey:
			hub.PushUnbondingMsg(entry.extra.(TimeAndSidWithAddr))
		case UnlockKey:
			hub.PushUnlockMsg( /*addr*/ entry.extra.(string), entry.bz)
		case CommentKey:
			hub.PushCommentMsg( /*token*/ entry.extra.(string), entry.bz)
		case CreateMarketInfoKey:
			hub.PushMarketInfoMsg( /*market*/ entry.extra.(string), entry.bz)
		case CreateOrderKey:
			hub.PushCreateOrderInfoMsg( /*addr*/ entry.extra.(string), entry.bz)
		case FillOrderKey:
			hub.PushFillOrderInfoMsg( /*addr*/ entry.extra.(string), entry.bz)
		case DealKey:
			hub.PushDealInfoMsg( /*market*/ entry.extra.(string), entry.bz)
		case CancelOrderKey:
			hub.PushCancelOrderMsg( /*addr*/ entry.extra.(string), entry.bz)
		case BancorTradeKey:
			hub.PushBancorTradeInfoMsg( /*addr*/ entry.extra.(string), entry.bz)
		case BancorDealKey:
			hub.PushBancorDealMsg( /*market*/ entry.extra.(string), entry.bz)
		case BancorKey:
			hub.PushBancorMsg( /*market*/ entry.extra.(string), entry.bz)
		case SlashKey:
			hub.PushSlashMsg(entry.bz)
		case TickerKey:
			hub.PushTickerMsg(entry.extra) // TODO. will modify param type
		case DepthFull:
			hub.PushDepthFullMsg( /*market*/ entry.extra.(string))
		case DepthKey:
			hub.PushDepthMsg( /*market*/ entry.bz, entry.extra.(map[string][]byte))
		case OptionKey:
			hub.subMan.SetSkipOption(entry.extra.(bool))
		case ValidatorCommissionKey:
			hub.PushValidatorCommissionMsg( /*addr*/ entry.extra.(string), entry.bz)
		case DelegationRewardsKey:
			hub.PushDelegationRewardsMsg( /*addr*/ entry.extra.(string), entry.bz)
		}
	}
}

func (hub *Hub) PushHeightInfoMsg(bz []byte) {
	infos := hub.subMan.GetHeightSubscribeInfo()
	for _, ss := range infos {
		hub.subMan.PushHeight(ss, bz)
	}
}

// TODO. Add test
func (hub *Hub) PushCandleMsg(market string, bz []byte, timeSpan string) {
	var targets []Subscriber
	info := hub.subMan.GetCandleStickSubscribeInfo()
	if info != nil {
		targets = info[market]
	}
	for _, target := range targets {
		ts, ok := target.Detail().(string)
		if !ok || ts != timeSpan {
			continue
		}
		hub.subMan.PushCandleStick(target, bz)
	}
}

func (hub *Hub) PushLockedCoinsMsg(addr string, bz []byte) {
	infos := hub.subMan.GetLockedSubscribeInfo()
	if conns, ok := infos[addr]; ok {
		for _, c := range conns {
			hub.subMan.PushLockedSendMsg(c, bz)
		}
	}
}

func (hub *Hub) PushIncomeMsg(receiver string, bz []byte) {
	info := hub.subMan.GetIncomeSubscribeInfo()
	targets, ok := info[receiver]
	if ok {
		for _, target := range targets {
			hub.subMan.PushIncome(target, bz)
		}
	}
}

func (hub *Hub) PushTxMsg(addr string, bz []byte) {
	info := hub.subMan.GetTxSubscribeInfo()
	targets, ok := info[addr]
	if ok {
		for _, target := range targets {
			hub.subMan.PushTx(target, bz)
		}
	}
}

func (hub *Hub) PushRedelegationMsg(param TimeAndSidWithAddr) {
	info := hub.subMan.GetRedelegationSubscribeInfo()
	targets, ok := info[param.addr]
	if ok {
		// query the redelegations whose completion time is between current block and last block
		end := hub.getEventKeyWithSidAndTime(RedelegationByte, param.addr, param.currTime, param.sid)
		start := hub.getEventKeyWithSidAndTime(RedelegationByte, param.addr, param.lastTime-1, param.sid)
		hub.dbMutex.RLock()
		iter := hub.db.ReverseIterator(start, end)
		defer func() {
			iter.Close()
			hub.dbMutex.RUnlock()
		}()
		for ; iter.Valid(); iter.Next() {
			for _, target := range targets {
				hub.subMan.PushRedelegation(target, iter.Value())
			}
		}
	}
}

func (hub *Hub) PushUnbondingMsg(param TimeAndSidWithAddr) {
	info := hub.subMan.GetUnbondingSubscribeInfo()
	targets, ok := info[param.addr]
	if ok {
		// query the unbondings whose completion time is between current block and last block
		end := hub.getEventKeyWithSidAndTime(UnbondingByte, param.addr, param.currTime, param.sid)
		start := hub.getEventKeyWithSidAndTime(UnbondingByte, param.addr, param.lastTime-1, param.sid)
		hub.dbMutex.RLock()
		iter := hub.db.ReverseIterator(start, end)
		defer func() {
			iter.Close()
			hub.dbMutex.RUnlock()
		}()
		for ; iter.Valid(); iter.Next() {
			for _, target := range targets {
				hub.subMan.PushUnbonding(target, iter.Value())
			}
		}
	}
}

func (hub *Hub) PushUnlockMsg(addr string, bz []byte) {
	info := hub.subMan.GetUnlockSubscribeInfo()
	targets, ok := info[addr]
	if ok {
		for _, target := range targets {
			hub.subMan.PushUnlock(target, bz)
		}
	}
}

func (hub *Hub) PushCommentMsg(token string, bz []byte) {
	info := hub.subMan.GetCommentSubscribeInfo()
	targets, ok := info[token]
	if ok {
		for _, target := range targets {
			hub.subMan.PushComment(target, bz)
		}
	}
}

func (hub *Hub) PushMarketInfoMsg(market string, bz []byte) {
	info := hub.subMan.GetMarketSubscribeInfo()
	targets, ok := info[market]
	if ok {
		for _, target := range targets {
			hub.subMan.PushCreateMarket(target, bz)
		}
	}
}

func (hub *Hub) PushCreateOrderInfoMsg(addr string, bz []byte) {
	//Push to subscribers
	info := hub.subMan.GetOrderSubscribeInfo()
	targets, ok := info[addr]
	if ok {
		for _, target := range targets {
			hub.subMan.PushCreateOrder(target, bz)
		}
	}
}

func (hub *Hub) PushFillOrderInfoMsg(addr string, bz []byte) {
	//Push to subscribers
	info := hub.subMan.GetOrderSubscribeInfo()
	targets, ok := info[addr]
	if ok {
		for _, target := range targets {
			hub.subMan.PushFillOrder(target, bz)
		}
	}
}

func (hub *Hub) PushDealInfoMsg(market string, bz []byte) {
	info := hub.subMan.GetDealSubscribeInfo()
	targets, ok := info[market]
	if ok {
		for _, target := range targets {
			hub.subMan.PushDeal(target, bz)
		}
	}
}

func (hub *Hub) PushCancelOrderMsg(addr string, bz []byte) {
	info := hub.subMan.GetOrderSubscribeInfo()
	targets, ok := info[addr]
	if ok {
		for _, target := range targets {
			hub.subMan.PushCancelOrder(target, bz)
		}
	}
}

func (hub *Hub) PushBancorTradeInfoMsg(addr string, bz []byte) {
	info := hub.subMan.GetBancorTradeSubscribeInfo()
	targets, ok := info[addr]
	if ok {
		for _, target := range targets {
			hub.subMan.PushBancorTrade(target, bz)
		}
	}
}

func (hub *Hub) PushBancorDealMsg(market string, bz []byte) {
	info := hub.subMan.GetBancorDealSubscribeInfo()
	targets, ok := info[market]
	if ok {
		for _, target := range targets {
			hub.subMan.PushBancorDeal(target, bz)
		}
	}
}

func (hub *Hub) PushBancorMsg(market string, bz []byte) {
	info := hub.subMan.GetBancorInfoSubscribeInfo()
	targets, ok := info[market]
	if ok {
		for _, target := range targets {
			hub.subMan.PushBancorInfo(target, bz)
		}
	}
}

func (hub *Hub) PushSlashMsg(bz []byte) {
	infos := hub.subMan.GetSlashSubscribeInfo()
	for _, ss := range infos {
		hub.subMan.PushSlash(ss, bz)
	}
}

func (hub *Hub) PushTickerMsg(msg interface{}) {
	tkMap := msg.(map[string]*Ticker)
	infos := hub.subMan.GetTickerSubscribeInfo()
	for _, subscriber := range infos {
		marketList := subscriber.Detail().(map[string]struct{})
		tickerList := make([]*Ticker, 0, len(marketList))
		for market := range marketList {
			if ticker, ok := tkMap[market]; ok {
				tickerList = append(tickerList, ticker)
			}
		}
		if len(tickerList) != 0 {
			hub.subMan.PushTicker(subscriber, tickerList)
		}
	}
}

// TODO, will check topic
func (hub *Hub) PushDepthFullMsg(market string) {
	info := hub.subMan.GetDepthSubscribeInfo()
	targets, ok := info[market]
	if !ok {
		return
	}
	for _, target := range targets {
		level := target.Detail().(string)
		if bz, err := hub.getDepthFullData(market, level, MaxCount); err == nil {
			hub.subMan.PushDepthFullMsg(target, bz)
		}
	}
}

// Get all the depth data about 'market' and merge them according to 'level'.
// When level is "all", it means no merging.
func (hub *Hub) getDepthFullData(market string, level string, count int) ([]byte, error) {
	var (
		bz  []byte
		err error
	)
	sell, buy := hub.QueryDepth(market, count)
	if level == "all" {
		bz, err = encodeDepthData(market, buy, sell)
		return bz, err
	}

	buyLevel := mergePrice(buy, level, true)
	sellLevel := mergePrice(sell, level, false)
	bz, err = encodeDepthLevel(market, buyLevel, sellLevel)
	return bz, err
}

// TODO, will replace string with interface
func (hub *Hub) PushDepthMsg(data []byte, levelsData map[string][]byte) {
	market := string(data)
	info := hub.subMan.GetDepthSubscribeInfo()
	targets, ok := info[market]
	if ok {
		for _, target := range targets {
			level := target.Detail().(string)
			if len(levelsData[level]) != 0 {
				if level == "all" {
					hub.subMan.PushDepthWithChange(target, levelsData[level])
				} else {
					hub.subMan.PushDepthWithDelta(target, levelsData[level])
				}
			}
		}
	}
}

func (hub *Hub) PushValidatorCommissionMsg(addr string, data []byte) {
	info := hub.subMan.GetValidatorCommissionInfo()
	targets, ok := info[addr]
	if !ok {
		return
	}

	for _, target := range targets {
		hub.subMan.PushValidatorCommissionInfo(target, data)
	}
}

func (hub *Hub) PushDelegationRewardsMsg(addr string, data []byte) {
	info := hub.subMan.GetDelegationRewards()
	targets, ok := info[addr]
	if !ok {
		return
	}

	for _, target := range targets {
		hub.subMan.PushDelegationRewards(target, data)
	}
}
