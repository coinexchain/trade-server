package server

type MsgWriter interface {
	WriteKV(k, v []byte) error
	Close() error
	String() string
}
