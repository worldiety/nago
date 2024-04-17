package gorilla

import (
	"github.com/gorilla/websocket"
	"sync"
)

type WebsocketChannel struct {
	conn      *websocket.Conn
	observers map[int]func(msg []byte) error
	mutex     sync.RWMutex
	hnd       int
}

func NewWebsocketChannel(conn *websocket.Conn) *WebsocketChannel {
	return &WebsocketChannel{conn: conn, observers: map[int]func(msg []byte) error{}}
}

func (w *WebsocketChannel) Loop() error {
	for {
		_, message, err := w.conn.ReadMessage()
		if err != nil {
			return err
		}

		if err := w.dispatch(message); err != nil {
			return err
		}
	}
}

func (w *WebsocketChannel) dispatch(buf []byte) error {
	w.mutex.RLock()
	defer w.mutex.RUnlock()

	for _, f := range w.observers {
		if err := f(buf); err != nil {
			return err
		}
	}

	return nil
}

func (w *WebsocketChannel) Subscribe(f func(msg []byte) error) (destroy func()) {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	w.hnd++
	myHandle := w.hnd
	w.observers[myHandle] = f

	return func() {
		w.mutex.Lock()
		defer w.mutex.Unlock()
		delete(w.observers, myHandle)
	}
}

func (w *WebsocketChannel) Publish(msg []byte) error {
	return w.conn.WriteMessage(websocket.TextMessage, msg)
}

// PublishLocal dispatches the buffer directly to currently all registered subscribers.
func (w *WebsocketChannel) PublishLocal(buf []byte) error {
	w.mutex.RLock()
	defer w.mutex.RUnlock()
	for _, f := range w.observers {
		if err := f(buf); err != nil {
			return err
		}
	}

	return nil
}
