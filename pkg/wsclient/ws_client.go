package wsclient

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
)

type Client struct {
	serverURL   string
	sendChan    chan interface{}
	receiveChan chan interface{}
}

type MessageType string

const (
	MTUndefined = ""
	MTFlyMap    = "fly_map"
	MTPos       = "pos"
	MTLog       = "log"
	MTCmd       = "cmd"
)

type Message struct {
	Type    MessageType
	Content []byte
}

func New(serverURL string) *Client {
	return &Client{
		serverURL:   serverURL,
		sendChan:    make(chan interface{}, 1),
		receiveChan: make(chan interface{}, 1),
	}
}

func (c *Client) SendMessage(message Message) {
	c.sendChan <- message
}

func (c *Client) ReceiveMessage(ctx context.Context) Message {
	select {
	case <-ctx.Done():
		return Message{}
	case msgInterface := <-c.receiveChan:
		msg, ok := (msgInterface).(Message)
		if !ok {
			return Message{}
		}
		return msg
	}
}

func (c *Client) Run(ctx context.Context) {
	logrus.Warnf("started websocket client")
	const reconnectInterval = 5 * time.Second
	timer := time.NewTimer(0)
	wsURL := "ws" + strings.TrimPrefix(c.serverURL, "http") + "drone/ws/"
	for {
		select {
		case <-ctx.Done():
			timer.Stop()
			logrus.Warnf("stopped websocket client")
			return
		case <-timer.C:
			func() {
				conn, _, err := websocket.DefaultDialer.DialContext(ctx, wsURL, nil)
				if err != nil {
					logrus.Error(fmt.Errorf("error connecting to server's web socket: %w", err))
					return
				}
				defer func() { _ = conn.Close() }()

				go c.receiveMessages(ctx, conn)
				go c.sendMessages(ctx, conn)
			}()
			timer.Reset(reconnectInterval)
		}
	}

}

func (c *Client) receiveMessages(ctx context.Context, conn *websocket.Conn) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			var msg Message
			if err := conn.ReadJSON(&msg); err != nil {
				logrus.Error(fmt.Errorf("error reading message from web socket: %w", err))
				return
			}
			select {
			case c.receiveChan <- msg:
			case <-time.After(200 * time.Millisecond):
				continue
			}

		}
	}
}

func (c *Client) sendMessages(ctx context.Context, conn *websocket.Conn) {
	for {
		select {
		case <-ctx.Done():
			return
		case msg := <-c.sendChan:
			if err := conn.WriteJSON(msg); err != nil {
				logrus.Error(fmt.Errorf("error writing message to web socket: %w", err))
				return
			}
		}
	}
}
