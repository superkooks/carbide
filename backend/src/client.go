package main

import (
	"errors"
	"fmt"
	"net"

	"github.com/gorilla/websocket"
)

// A client represented a user that is currently connected via websocket
type Client struct {
	conn   *websocket.Conn
	events chan []byte
}

func CreateClient(c *websocket.Conn) *Client {
	return &Client{
		conn:   c,
		events: make(chan []byte, 16),
	}
}

// Listen for messages from the client
func (c *Client) Listen() {
	for {
		_, b, err := c.conn.ReadMessage()
		if err != nil {
			panic(err)
		}

		// Parse message type
		switch b[0] {
		case EVT_TYPE_DATA:
			fmt.Println("Data")
		case EVT_TYPE_SUBSCRIBE_GUILDS:
			fmt.Println("sub guilds")
		case EVT_TYPE_ADD_USER:
			fmt.Println("add user")
		default:
			panic("unsupported opcode")
		}
	}
}

// Listen for messages to the client
func (c *Client) Send() {
	for {
		e := <-c.events
		err := c.conn.WriteMessage(websocket.TextMessage, e)
		if errors.Is(err, net.ErrClosed) || errors.Is(err, websocket.ErrCloseSent) {
			return
		} else if err != nil {
			panic(err)
		}
	}
}
