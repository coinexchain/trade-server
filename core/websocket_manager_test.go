package core

import (
	"encoding/json"
	"sync"
	"testing"

	"github.com/gorilla/websocket"

	"github.com/stretchr/testify/require"
)

func TestGroupOfDataPacket(t *testing.T) {
	bz := groupOfDataPacket("test", nil)
	require.EqualValues(t, "{\"type\":\"test\", \"payload\":[]}", bz)

	data := make([]json.RawMessage, 2)
	data[0] = json.RawMessage("hello")
	data[1] = json.RawMessage("world")
	bz = groupOfDataPacket("test", data)
	require.EqualValues(t, "{\"type\":\"test\", \"payload\":[hello,world]}", bz)

}

func TestCheckTopicValid(t *testing.T) {
	assertZeroParamTopic(t)
	assertOneParamsTopic(t)
	assertTicker(t)
	assertDepth(t)
	assetKline(t)
}

func assertZeroParamTopic(t *testing.T) {
	// blockinfo
	require.True(t, checkTopicValid(BlockInfoKey, []string{}))
	require.False(t, checkTopicValid(BlockInfoKey, []string{"height"}))
	require.False(t, checkTopicValid(BlockInfoKey, []string{"height", "time"}))

	// slash
	require.True(t, checkTopicValid(SlashKey, []string{}))
	require.False(t, checkTopicValid(SlashKey, []string{"time"}))
}

func assertTicker(t *testing.T) {
	// ticker:abc/cet; ticker:B:abc/cet
	require.False(t, checkTopicValid(TickerKey, []string{}))
	require.True(t, checkTopicValid(TickerKey, []string{"abc"}))
	require.True(t, checkTopicValid(TickerKey, []string{"abc/cet"}))

	require.False(t, checkTopicValid(TickerKey, []string{"abc", "cet"}))
	require.True(t, checkTopicValid(TickerKey, []string{"B", "cet"}))
	require.True(t, checkTopicValid(TickerKey, []string{"B", "abc/cet"}))

	require.False(t, checkTopicValid(TickerKey, []string{"abc", "cet", "btc"}))
}

func assertOneParamsTopic(t *testing.T) {
	topics := []string{UnbondingKey, RedelegationKey, LockedKey,
		UnlockKey, TxKey, IncomeKey, OrderKey, CommentKey,
		BancorTradeKey, BancorKey, DealKey, BancorDealKey}

	//deal:<trading-pair>; income:<address> ...
	// ../docs/websocket-streams.md
	for _, topic := range topics {
		require.False(t, checkTopicValid(topic, []string{"topic-param1", "topic-param2"}))
	}
	for _, topic := range topics {
		require.True(t, checkTopicValid(topic, []string{"topic-param"}))
	}
}

func assetKline(t *testing.T) {
	// kline:abc/cet:1min; kline:B:abc/cet:1min

	require.True(t, checkTopicValid(KlineKey, []string{"abc/cet", "1min"}))
	require.True(t, checkTopicValid(KlineKey, []string{"abc/cet", "1hour"}))
	require.True(t, checkTopicValid(KlineKey, []string{"abc/cet", "1day"}))
	require.True(t, checkTopicValid(KlineKey, []string{"B", "abc/cet", "1min"}))
	require.True(t, checkTopicValid(KlineKey, []string{"B", "abc/cet", "1hour"}))
	require.True(t, checkTopicValid(KlineKey, []string{"B", "abc/cet", "1day"}))

	//invalid timespan
	require.False(t, checkTopicValid(KlineKey, []string{"abc/cet", "1week"}))
	require.False(t, checkTopicValid(KlineKey, []string{"B", "abc/cet", "1week"}))

	// invalid params number
	require.False(t, checkTopicValid(KlineKey, []string{}))
	require.False(t, checkTopicValid(KlineKey, []string{"param"}))
	require.False(t, checkTopicValid(KlineKey, []string{"param1", "param2", "param3", "param4"}))
}

func assertDepth(t *testing.T) {
	// depth:<trading-pair>:<level>
	levels := []string{"100", "10", "1", "0.1", "0.01", "0.001",
		"0.0001", "0.00001", "0.000001",
		"0.0000001", "0.00000001", "0.000000001", "0.0000000001",
		"0.00000000001", "0.000000000001", "0.0000000000001",
		"0.00000000000001", "0.000000000000001", "0.0000000000000001",
		"0.00000000000000001", "0.000000000000000001", "all"}
	for _, level := range levels {
		require.True(t, checkTopicValid(DepthKey, []string{"abc/cet", level}))
	}

	// Backward compatibility;
	// Since the first version did not include level, so if you pass only
	// one param in depth subscribe, the default level is "all".
	require.True(t, checkTopicValid(DepthKey, []string{"abc/cet"}))

	// invalid param number
	require.False(t, checkTopicValid(DepthKey, []string{}))
	require.False(t, checkTopicValid(DepthKey, []string{"param1", "param2", "param3"}))

	// invalid level
	require.False(t, checkTopicValid(DepthKey, []string{"abc/cet", "invalid-level"}))
}

func TestWebsocketManager_AddSubscribeConnAndRemoveSubscribeConn(t *testing.T) {
	wg := sync.WaitGroup{}
	for i := 0; i < 10; i++ {
		var (
			wsManager = NewWebSocketManager()
			c1        = NewConn(&websocket.Conn{})
			c2        = NewConn(&websocket.Conn{})
		)
		wg.Add(2)
		//go assertAddSubscribeConn(t, &wg, wsManager, c1, c2)
		//go assertRemoveSubscribeConn(t, &wg, wsManager, c1, c2)
		assertAddSubscribeConn(t, &wg, wsManager, c1, c2)
		assertRemoveSubscribeConn(t, &wg, wsManager, c1, c2)
	}
	wg.Wait()
}

func assertAddSubscribeConn(t *testing.T, wg *sync.WaitGroup, wsManager *WebsocketManager, c1, c2 *Conn) {
	defer wg.Done()
	// zero param: blockinfo
	wsManager.AddSubscribeConn(c1, SlashKey, nil)
	wsManager.AddSubscribeConn(c1, BlockInfoKey, nil)
	require.EqualValues(t, 2, len(wsManager.topics2Conns))
	require.EqualValues(t, 2, len(c1.allTopics))
	require.EqualValues(t, 0, len(c1.topicWithParams))

	// ticker:abc/cet; ticker:B:abc/cet
	wsManager.AddSubscribeConn(c2, TickerKey, []string{"abc/cet"})
	wsManager.AddSubscribeConn(c2, TickerKey, []string{"B", "abc/cet"})
	require.EqualValues(t, 3, len(wsManager.topics2Conns))
	require.EqualValues(t, 1, len(c2.allTopics))
	require.EqualValues(t, 2, len(c2.topicWithParams[TickerKey]))
}

func assertRemoveSubscribeConn(t *testing.T, wg *sync.WaitGroup, wsManager *WebsocketManager, c1, c2 *Conn) {
	defer wg.Done()
	// zero param: blockinfo
	wsManager.RemoveSubscribeConn(c1, SlashKey, nil)
	require.EqualValues(t, 2, len(wsManager.topics2Conns))
	wsManager.RemoveSubscribeConn(c1, BlockInfoKey, nil)
	require.EqualValues(t, 0, len(c1.allTopics))
	require.EqualValues(t, 0, len(c1.topicWithParams))
	require.EqualValues(t, 1, len(wsManager.topics2Conns))

	// ticker:abc/cet; ticker:B:abc/cet
	wsManager.RemoveSubscribeConn(c2, TickerKey, []string{"abc/cet"})
	require.EqualValues(t, 1, len(c2.allTopics))
	require.EqualValues(t, 1, len(wsManager.topics2Conns))
	require.EqualValues(t, 1, len(c2.topicWithParams))
	expectVal := make(map[string]struct{})
	expectVal["B:abc/cet"] = struct{}{}
	require.EqualValues(t, expectVal, c2.topicWithParams[TickerKey])

	wsManager.RemoveSubscribeConn(c2, TickerKey, []string{"B", "abc/cet"})
	require.EqualValues(t, 0, len(c2.allTopics))
	require.EqualValues(t, 0, len(wsManager.topics2Conns))
	require.EqualValues(t, 0, len(c2.topicWithParams))
}

func TestWebsocketManager_AddWsConnAndCloseConn_MulThread(t *testing.T) {
	var (
		wg        sync.WaitGroup
		wsManager = NewWebSocketManager()
	)
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go assertAddAndCloseConn(&wg, wsManager)
	}
	wg.Wait()
}

func assertAddAndCloseConn(wg *sync.WaitGroup, wsManager *WebsocketManager) {
	defer wg.Done()
	c1 := wsManager.AddWsConn(&mockWsConn{})
	wsManager.CloseWsConn(c1)
}

func TestWebsocketManager_AddWsConnAndCloseConn_SigThread(t *testing.T) {
	var wsManager = NewWebSocketManager()
	c1 := wsManager.AddWsConn(&mockWsConn{})
	require.EqualValues(t, 1, len(wsManager.wsConn2Conn))
	wsManager.CloseWsConn(c1)
	require.EqualValues(t, 0, len(wsManager.wsConn2Conn))
}
