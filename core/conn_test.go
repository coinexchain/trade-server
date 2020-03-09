package core

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type MockWif struct {
	mtx    sync.Mutex
	recode [][]byte
}

func (m *MockWif) Close() error {
	return nil
}

func (m *MockWif) WriteMessage(msgType int, data []byte) error {
	m.mtx.Lock()
	defer m.mtx.Unlock()
	m.recode = append(m.recode, data)
	return nil
}

func (m *MockWif) ReadMessage() (messageType int, p []byte, err error) {
	panic("implement me")
}

func (m *MockWif) SetPingHandler(func(string) error) {

}

func (m *MockWif) PingHandler() func(string) error {
	panic("implement me")
}

func (m *MockWif) GetRecords() [][]byte {
	m.mtx.Lock()
	defer m.mtx.Unlock()
	ret := make([][]byte, len(m.recode))
	copy(ret, m.recode)
	return ret
}

func TestSendMsg(t *testing.T) {
	mockWsIf := MockWif{}
	c := NewConn(&mockWsIf)
	for i := 0; i < 5; i++ {
		c.WriteMsg([]byte(fmt.Sprintf("%d", i)))
	}
	time.Sleep(time.Second * 2)
	c.Close()

	records := mockWsIf.GetRecords()
	require.EqualValues(t, 5, len(records))
	for i := 0; i < 5; i++ {
		require.EqualValues(t, fmt.Sprintf("%d", i), string(records[i]))
	}
}
