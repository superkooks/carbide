package main

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"net"
	"os"
	"syscall"

	"github.com/google/uuid"
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
		if errors.Is(err, net.ErrClosed) || errors.Is(err, syscall.ECONNRESET) || os.IsTimeout(err) || websocket.IsCloseError(err, 1000, 1001, 1005, 1006) {
			fmt.Println("client lost connection")
			return
		} else if err != nil {
			panic(err)
		}

		// Parse message type
		switch b[0] {
		case EVT_TYPE_DATA:
			fmt.Println("Data")

		case EVT_TYPE_SUBSCRIBE_GUILDS:
			fmt.Println("sub guilds")
			fmt.Println("mesg len", len(b))
			fmt.Println(hex.EncodeToString(b))

			var count int64
			binary.Read(bytes.NewBuffer(b[1:]), binary.BigEndian, &count)
			fmt.Println(count)

			var uuids []uuid.UUID
			for i := int64(0); i < count; i++ {
				var u uuid.UUID
				copy(u[:], b[9+i*16:25+i*16])
				uuids = append(uuids, u)
			}

			fmt.Println(uuids)

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
