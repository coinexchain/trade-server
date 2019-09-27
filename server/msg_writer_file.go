package server

import (
	"fmt"
	"io"
	"os"
)

type fileMsgWriter struct {
	io.WriteCloser
}

func NewFileMsgWriter(filePath string) (MsgWriter, error) {
	file, err := openFile(filePath)
	if err != nil {
		return nil, err
	}
	return fileMsgWriter{file}, nil
}

func openFile(filePath string) (*os.File, error) {
	if s, err := os.Stat(filePath); os.IsNotExist(err) {
		return os.Create(filePath)
	} else if s.IsDir() {
		return nil, fmt.Errorf("Need to give the file path ")
	} else {
		return os.OpenFile(filePath, os.O_RDWR|os.O_APPEND, 0666)
	}
}

func (w fileMsgWriter) WriteKV(k, v []byte) error {
	if _, err := w.WriteCloser.Write(k); err != nil {
		return err
	}
	if _, err := w.WriteCloser.Write([]byte("#")); err != nil {
		return err
	}
	if _, err := w.WriteCloser.Write(v); err != nil {
		return err
	}
	if _, err := w.WriteCloser.Write([]byte("\n")); err != nil {
		return err
	}

	return nil
}

func (w fileMsgWriter) Close() error {
	return w.WriteCloser.Close()
}

func (w fileMsgWriter) String() string {
	return "file"
}
