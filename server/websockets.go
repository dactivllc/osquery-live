package main

import (
	"github.com/gorilla/websocket"
	"github.com/pkg/errors"
)

type websocketWrapper struct {
	conn *websocket.Conn
}

func (w *websocketWrapper) Read(p []byte) (int, error) {
	_, reader, err := w.conn.NextReader()
	if err != nil {
		return 0, errors.Wrap(err, "get websocket reader")
	}

	n, err := reader.Read(p)
	if err != nil {
		return n, errors.Wrap(err, "read from websocket reader")
	}

	return n, nil
}

func (w *websocketWrapper) Write(p []byte) (int, error) {
	writer, err := w.conn.NextWriter(websocket.TextMessage)
	if err != nil {
		return 0, errors.Wrap(err, "get websocket writer")
	}

	n, err := writer.Write(p)
	if err != nil {
		return 0, errors.Wrap(err, "write to websocket writer")
	}

	if err := writer.Close(); err != nil {
		return n, errors.Wrap(err, "close websocket writer")
	}

	return n, nil
}
