package core

type mockWsConn struct {
	num int
}

func (m *mockWsConn) Close() error {
	return nil
}

func (m *mockWsConn) WriteMessage(msgType int, data []byte) error {
	panic("implement me")
}

func (m *mockWsConn) ReadMessage() (messageType int, p []byte, err error) {
	panic("implement me")
}

func (m *mockWsConn) SetPingHandler(func(string) error) {
}

func (m *mockWsConn) PingHandler() func(string) error {
	panic("implement me")
}
