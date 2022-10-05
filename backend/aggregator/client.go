package main

import (
	"carbide/backend/common"
	"crypto/rand"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/vmihailenco/msgpack"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

//lint:file-ignore SA1012 Mongo permits it

// A client represented a user that is currently connected via websocket
type Client struct {
	Conn   *websocket.Conn
	Events chan []byte

	LastAck    time.Time
	Authorized bool
	UserID     uuid.UUID // not set until Authorized is true
}

var guilds = make(map[uuid.UUID][]*Client)

func NewClient(c *websocket.Conn) *Client {
	return &Client{
		Conn:    c,
		Events:  make(chan []byte, 16),
		LastAck: time.Now(),
	}
}

// Listen for messages from the client
func (c *Client) Listen() {
	for {
		_, b, err := c.Conn.ReadMessage()
		if err != nil {
			fmt.Println("client lost connection")

			// Remove the client's subscription from all guilds
			for guildID := range guilds {
				// A potential optimisation would be to store the guilds the user is subscribed to
				// in order to speed up this search
				guilds[guildID] = common.RemoveFromSlice(guilds[guildID], c)

				// TODO if this slice is empty, we should probably tell the reflector as such
			}
			return
		}

		var event common.Event
		err = msgpack.Unmarshal(b, &event)
		if err != nil {
			panic(err)
		}

		if !c.Authorized && (event.Type != common.EVT_AUTHENTICATE && event.Type != common.EVT_REGISTER) {
			// Clients must authenticate first
			c.Close(4000, "failed to authenticate")
			return
		}

		// Parse message type
		switch event.Type {
		case common.EVT_HEARTBEAT_ACK:
			c.LastAck = time.Now()

		case common.EVT_DATA:
			var evt common.EvtData
			err = msgpack.Unmarshal(event.Evt, &evt)
			if err != nil {
				panic(err)
			}

			// TODO Check whether user is authorized to send message to that guild

			// Send the message to the appropriate reflector
			ref := ReflectorForGuild(evt.GuildID)
			ref.Events <- b

		case common.EVT_REGISTER:
			// Generate user id and token
			user := common.DBUser{
				ID:    uuid.New(),
				Token: make([]byte, 16),
			}
			rand.Read(user.Token)

			// Add to database
			_, err = db.Collection("users").InsertOne(nil, user)
			if err != nil {
				panic(err)
			}

			// Set connection credentials
			c.Authorized = true
			c.UserID = user.ID

			// Reply with Register
			c.Events <- common.WrapEvent(common.EVT_REGISTER, common.EvtRegister{
				UserID: user.ID,
				Token:  user.Token,
			})

		case common.EVT_AUTHENTICATE:
			var evt common.EvtAuthenticate
			err = msgpack.Unmarshal(event.Evt, &evt)
			if err != nil {
				panic(err)
			}

			// Check whether token matches any in the db
			var user common.DBUser
			err = db.Collection("users").FindOne(nil, bson.D{{Key: "token", Value: evt.Token}}).Decode(&user)
			if err == mongo.ErrNoDocuments {
				// If there are no results, that means the token does not exist,
				// so you will be disconnected
				c.Close(4000, "failed to authenticate")
				return
			} else if err != nil {
				panic(err)
			}

			// Set connection credentials
			c.Authorized = true
			c.UserID = user.ID

		case common.EVT_SUB_GUILDS:
			var evt common.EvtSubGuilds
			err = msgpack.Unmarshal(event.Evt, &evt)
			if err != nil {
				panic(err)
			}

			for _, guildID := range evt.GuildIDs {
				// TODO Check whether user is authorized to subscribe to that guild

				// Send single sub guild to reflector
				// NOTE: we could optimise this to not send the message if we are already
				//       subscribed, but that sounds like unecessary effort
				// NOTE: we could also optimise this to group the guilds per reflector, but that
				//       also sounds like effort
				ref := ReflectorForGuild(guildID)
				ref.Events <- common.WrapEvent(common.EVT_SUB_GUILDS, common.EvtSubGuilds{
					GuildIDs: []uuid.UUID{guildID},
				})

				// Add the user to the list of who we notify when we receive a message for a guild
				clients, ok := guilds[guildID]
				if !ok {
					clients = []*Client{}
				}
				clients = append(clients, c)
				guilds[guildID] = clients
			}

		case common.EVT_ADD_USERS:

		case common.EVT_REMOVE_USERS:

		default:
			panic("unsupported event type")
		}
	}
}

// Listen for messages to the client
func (c *Client) Send() {
	for {
		evt := <-c.Events

		c.Conn.SetWriteDeadline(time.Now().Add(time.Second))
		err := c.Conn.WriteMessage(websocket.BinaryMessage, evt)
		if err != nil {
			return
		}
	}
}

// Send hearbeats every interval, checking whether the connection has failed
func (c *Client) Heartbeat() {
	for {
		time.Sleep(common.CLIENT_HEARTBEAT_INTERVAL)

		if time.Since(c.LastAck) > 2*common.CLIENT_HEARTBEAT_INTERVAL {
			// Close the connection if twice the heartbeat interval has elapsed
			c.Close(4001, "failed to ack heartbeat")
			return
		}

		c.Events <- common.WrapEvent(common.EVT_HEARTBEAT, nil)
	}
}

// Close is a helper for closing the connection with a close code
func (c *Client) Close(code int, msg string) {
	c.Conn.WriteControl(
		websocket.CloseMessage,
		websocket.FormatCloseMessage(code, msg),
		time.Now().Add(time.Second))
	time.Sleep(time.Second)
	c.Conn.Close()
}
