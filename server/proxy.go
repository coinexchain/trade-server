package server

import "github.com/coinexchain/trade-server/core"

// TODO: delete later
type TestSubscriber struct {
}

func (t TestSubscriber) Detail() interface{} {
	return []string{"detail"}
}

func (t TestSubscriber) WriteMsg([]byte) error {
	return nil
}

type TestSubscribeManager struct {
}

func (sm TestSubscribeManager) GetSlashSubscribeInfo() []core.Subscriber {
	return []core.Subscriber{TestSubscriber{}}
}

func (sm TestSubscribeManager) GetHeightSubscribeInfo() []core.Subscriber {
	return []core.Subscriber{TestSubscriber{}}
}

func (sm TestSubscribeManager) GetTickerSubscribeInfo() []core.Subscriber {
	return []core.Subscriber{TestSubscriber{}}
}

func (sm TestSubscribeManager) GetCandleStickSubscribeInfo() map[string][]core.Subscriber {
	return map[string][]core.Subscriber{"test": {TestSubscriber{}}}
}

func (sm TestSubscribeManager) GetDepthSubscribeInfo() map[string][]core.Subscriber {
	return map[string][]core.Subscriber{"test": {TestSubscriber{}}}
}

func (sm TestSubscribeManager) GetDealSubscribeInfo() map[string][]core.Subscriber {
	return map[string][]core.Subscriber{"test": {TestSubscriber{}}}
}

func (sm TestSubscribeManager) GetBancorInfoSubscribeInfo() map[string][]core.Subscriber {
	return map[string][]core.Subscriber{"test": {TestSubscriber{}}}
}

func (sm TestSubscribeManager) GetCommentSubscribeInfo() map[string][]core.Subscriber {
	return map[string][]core.Subscriber{"test": {TestSubscriber{}}}
}

func (sm TestSubscribeManager) GetOrderSubscribeInfo() map[string][]core.Subscriber {
	return map[string][]core.Subscriber{"test": {TestSubscriber{}}}
}

func (sm TestSubscribeManager) GetBancorTradeSubscribeInfo() map[string][]core.Subscriber {
	return map[string][]core.Subscriber{"test": {TestSubscriber{}}}
}
func (sm TestSubscribeManager) GetIncomeSubscribeInfo() map[string][]core.Subscriber {
	return map[string][]core.Subscriber{"test": {TestSubscriber{}}}
}
func (sm TestSubscribeManager) GetUnbondingSubscribeInfo() map[string][]core.Subscriber {
	return map[string][]core.Subscriber{"test": {TestSubscriber{}}}
}
func (sm TestSubscribeManager) GetRedelegationSubscribeInfo() map[string][]core.Subscriber {
	return map[string][]core.Subscriber{"test": {TestSubscriber{}}}
}
func (sm TestSubscribeManager) GetUnlockSubscribeInfo() map[string][]core.Subscriber {
	return map[string][]core.Subscriber{"test": {TestSubscriber{}}}
}
func (sm TestSubscribeManager) GetTxSubscribeInfo() map[string][]core.Subscriber {
	return map[string][]core.Subscriber{"test": {TestSubscriber{}}}
}

func (sm TestSubscribeManager) PushSlash(subscriber core.Subscriber, info []byte)       {}
func (sm TestSubscribeManager) PushHeight(subscriber core.Subscriber, info []byte)      {}
func (sm TestSubscribeManager) PushTicker(subscriber core.Subscriber, t []*core.Ticker) {}
func (sm TestSubscribeManager) PushDepthSell(subscriber core.Subscriber, info []byte)   {}

func (sm TestSubscribeManager) PushDepthBuy(subscriber core.Subscriber, info []byte) {}

func (sm TestSubscribeManager) PushCandleStick(subscriber core.Subscriber, info []byte)  {}
func (sm TestSubscribeManager) PushDeal(subscriber core.Subscriber, info []byte)         {}
func (sm TestSubscribeManager) PushCreateOrder(subscriber core.Subscriber, info []byte)  {}
func (sm TestSubscribeManager) PushFillOrder(subscriber core.Subscriber, info []byte)    {}
func (sm TestSubscribeManager) PushCancelOrder(subscriber core.Subscriber, info []byte)  {}
func (sm TestSubscribeManager) PushBancorInfo(subscriber core.Subscriber, info []byte)   {}
func (sm TestSubscribeManager) PushBancorTrade(subscriber core.Subscriber, info []byte)  {}
func (sm TestSubscribeManager) PushIncome(subscriber core.Subscriber, info []byte)       {}
func (sm TestSubscribeManager) PushUnbonding(subscriber core.Subscriber, info []byte)    {}
func (sm TestSubscribeManager) PushRedelegation(subscriber core.Subscriber, info []byte) {}
func (sm TestSubscribeManager) PushUnlock(subscriber core.Subscriber, info []byte)       {}
func (sm TestSubscribeManager) PushTx(subscriber core.Subscriber, info []byte)           {}
func (sm TestSubscribeManager) PushComment(subscriber core.Subscriber, info []byte)      {}
