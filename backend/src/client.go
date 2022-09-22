package main

import (
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
	Conn   *websocket.Conn
	Events chan []byte

	Guilds []uuid.UUID
}

func CreateClient(c *websocket.Conn) *Client {
	return &Client{
		Conn:   c,
		Events: make(chan []byte, 16),
	}
}

// Listen for messages from the client
func (c *Client) Listen() {
	for {
		_, b, err := c.Conn.ReadMessage()
		if errors.Is(err, net.ErrClosed) || errors.Is(err, syscall.ECONNRESET) || os.IsTimeout(err) || websocket.IsCloseError(err, 1000, 1001, 1005, 1006) {
			fmt.Println("client lost connection")

			// Remove the client's subscription
			// NOTE: There has got to be a better way to do this
			for _, g := range c.Guilds {
				for k, cl := range guilds[g].Clients {
					if cl == c {
						guilds[g].Clients = append(guilds[g].Clients[:k], guilds[g].Clients[k+1:]...)
					}
				}
			}

			return
		} else if err != nil {
			panic(err)
		}

		// Parse message type
		switch b[0] {
		case EVT_TYPE_DATA:
			fmt.Println("Data")

			// Hmmm, these offsets will never cause any pain
			var guild uuid.UUID
			copy(guild[:], b[1:17])

			// TODO: Store message

			// Send message to clients in guild
			for _, cl := range guilds[guild].Clients {
				cl.Events <- b
			}

		case EVT_TYPE_SUBSCRIBE_GUILDS:
			fmt.Println("sub guilds")
			fmt.Println("mesg len", len(b))
			fmt.Println(hex.EncodeToString(b))

			// This is awful
			var uuids []uuid.UUID
			for i := 0; i < (len(b)-1)/16; i++ {
				var u uuid.UUID
				copy(u[:], b[1+i*16:17+i*16])
				uuids = append(uuids, u)

				// TODO: Check whether user has been added to guild

				// Subscribe client to requested guilds
				if _, ok := guilds[u]; !ok {
					guilds[u] = new(Guild)
				}
				guilds[u].Clients = append(guilds[u].Clients, c)
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
		e := <-c.Events
		err := c.Conn.WriteMessage(websocket.BinaryMessage, e)
		if errors.Is(err, net.ErrClosed) || errors.Is(err, websocket.ErrCloseSent) {
			return
		} else if err != nil {
			panic(err)
		}
	}
}
