package core

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type MockWif struct {
	recode [][]byte
}

func (m *MockWif) Close() error {
	panic("implement me")
}

func (m *MockWif) WriteMessage(msgType int, data []byte) error {
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

func TestSendMsg(t *testing.T) {
	mockWsIf := MockWif{}
	c := NewConn(&mockWsIf)

	go func() {
		c.sendMsg()
	}()
	for i := 0; i < 5; i++ {
		c.WriteMsg([]byte(fmt.Sprintf("%d", i)))
	}
	time.Sleep(time.Second * 3)
	close(c.msgChan)

	require.EqualValues(t, 5, len(mockWsIf.recode))
	for i := 0; i < 5; i++ {
		require.EqualValues(t, fmt.Sprintf("%d", i), string(mockWsIf.recode[i]))
	}
}
