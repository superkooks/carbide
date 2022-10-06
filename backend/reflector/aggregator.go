package main

import (
	"carbide/backend/common"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/vmihailenco/msgpack"
)

//lint:file-ignore SA1012 Mongo permits it

type Aggregator struct {
	Conn   *websocket.Conn
	Events chan []byte

	LastAck time.Time
}

var guilds = make(map[uuid.UUID][]*Aggregator)

func NewAggregator(conn *websocket.Conn) *Aggregator {
	return &Aggregator{
		Conn:    conn,
		Events:  make(chan []byte, 16),
		LastAck: time.Now(),
	}
}

func (a *Aggregator) Listen() {
	for {
		_, b, err := a.Conn.ReadMessage()
		if err != nil {
			fmt.Println("aggregator lost connection")

			// Retry, otherwise remove aggregator from list
			// Note: we need to think about how heartbeats and retries should work
			//       and how connections can fail
			// TODO Remove the aggregator from the list
			return
		}

		var event common.Event
		err = msgpack.Unmarshal(b, &event)
		if err != nil {
			panic(err)
		}

		switch event.Type {
		case common.EVT_HEARTBEAT_ACK:
			a.LastAck = time.Now()

		case common.EVT_DATA:
			var evt common.EvtData
			err = msgpack.Unmarshal(event.Evt, &evt)
			if err != nil {
				panic(err)
			}

			evt.Timestamp = time.Now().UnixMilli()

			// Store the message
			m := common.DBMessage{
				ID:        evt.EvtID,
				GuildID:   evt.GuildID,
				Timestamp: evt.Timestamp,
				Message:   evt.Message,
			}
			_, err = db.Collection("messages").InsertOne(nil, m)
			if err != nil {
				panic(err)
			}

			// Reflect the messages to the appropriate aggregators
			for _, v := range guilds[evt.GuildID] {
				v.Events <- b
			}

		case common.EVT_SUB_GUILDS:
			var evt common.EvtSubGuilds
			err = msgpack.Unmarshal(event.Evt, &evt)
			if err != nil {
				panic(err)
			}

			// The sub guild data is vetted by the aggregator, so we can trust it
			for _, v := range evt.GuildIDs {
				aggs := guilds[v]
				aggs = append(aggs, a)
				guilds[v] = aggs
			}
		}

	}
}

func (a *Aggregator) Heartbeat() {
	for {
		time.Sleep(common.INTERNAL_HEARTBEAT_INTERVAL)

		if time.Since(a.LastAck) > 2*common.INTERNAL_HEARTBEAT_INTERVAL {
			// Close the connection if twice the heartbeat interval has elapsed
			a.Close(4001, "failed to ack heartbeat")
			return
		}

		a.Events <- common.WrapEvent(common.EVT_HEARTBEAT, nil)
	}
}

// Close is a helper for closing the connection with a close code
func (a *Aggregator) Close(code int, msg string) {
	a.Conn.WriteControl(
		websocket.CloseMessage,
		websocket.FormatCloseMessage(code, msg),
		time.Now().Add(time.Second))
	time.Sleep(time.Second)
	a.Conn.Close()
}

func (a *Aggregator) Send() {
	for {
		evt := <-a.Events
		messagesRelayed++

		a.Conn.SetWriteDeadline(time.Now().Add(time.Second))
		err := a.Conn.WriteMessage(websocket.BinaryMessage, evt)
		if err != nil {
			return
		}
	}
}
