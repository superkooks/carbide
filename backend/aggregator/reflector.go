package main

import (
	"fmt"
	"sort"
	"time"

	"carbide/backend/common"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/vmihailenco/msgpack"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

//lint:file-ignore SA1012 Mongo permits it

type Reflector struct {
	Addr   string
	Conn   *websocket.Conn
	Events chan []byte

	LastBeat time.Time
}

var reflectors []*Reflector

func ReflectorForGuild(guildID uuid.UUID) *Reflector {
	fmt.Println("finding reflector for guild", guildID.String())

	// Find the guild object
	var guild common.DBGuild
	err := db.Collection("guilds").
		FindOne(nil, bson.D{{Key: "_id", Value: guildID}}).Decode(&guild)
	if err == mongo.ErrNoDocuments {
		// Insert guild if it doesn't exist
		guild = common.DBGuild{
			ID: guildID,
		}

		_, err = db.Collection("guilds").InsertOne(nil, guild)
		if err != nil {
			panic(err)
		}
	} else if err != nil {
		panic(err)
	}

	// If a reflector hasn't been allocated, then allocate one
	if guild.ReflectorAddr == "" {
		// Get all reflectors
		var reflectors []common.DBReflector
		cur, err := db.Collection("reflectors").
			Find(nil, bson.D{})
		if err != nil {
			panic(err)
		}
		err = cur.All(nil, &reflectors)
		if err != nil {
			panic(err)
		}

		if len(reflectors) == 0 {
			panic("no reflectors to allocate to")
		}

		// Sort for lowest load
		sort.Sort(common.ByLoad(reflectors))

		// Update guild object
		fmt.Println("chosen lowest load reflector:", reflectors[0].Addr)
		guild.ReflectorAddr = reflectors[0].Addr
		_, err = db.Collection("guilds").
			ReplaceOne(nil, bson.D{{Key: "_id", Value: guildID}}, guild)
		if err != nil {
			panic(err)
		}
	}

	fmt.Println("using reflector:", guild.ReflectorAddr)
	return FindReflector(guild.ReflectorAddr)
}

func FindReflector(addr string) *Reflector {
	for _, ref := range reflectors {
		if ref.Addr == addr {
			return ref
		}
	}

	// If we didn't find the reflector, connect to it
	return newReflector(addr)
}

func newReflector(addr string) *Reflector {
	// Note: we should probably use wss between machines, but...
	conn, _, err := websocket.DefaultDialer.Dial("ws://"+addr, nil)
	if err != nil {
		panic(err)
	}

	r := &Reflector{
		Addr:     addr,
		Conn:     conn,
		Events:   make(chan []byte, 16),
		LastBeat: time.Now(),
	}

	go r.Watchdog()
	go r.Listen()
	go r.Send()

	reflectors = append(reflectors, r)
	return r
}

func (r *Reflector) Listen() {
	for {
		_, b, err := r.Conn.ReadMessage()
		if err != nil {
			fmt.Println("client lost connection")

			// TODO Remove the reflector from the list

			return
		}

		var event common.Event
		err = msgpack.Unmarshal(b, &event)
		if err != nil {
			panic(err)
		}

		switch event.Type {
		case common.EVT_HEARTBEAT:
			// Respond with heartbeat ack
			r.LastBeat = time.Now()
			r.Events <- common.WrapEvent(common.EVT_HEARTBEAT_ACK, nil)

		case common.EVT_DATA:
			var evt common.EvtData
			err = msgpack.Unmarshal(event.Evt, &evt)
			if err != nil {
				panic(err)
			}

			// Distribute this events to all subscribed clients
			for _, v := range guilds[evt.GuildID] {
				v.Events <- b
			}
		}

	}
}

func (r *Reflector) Watchdog() {
	for {
		time.Sleep(common.INTERNAL_HEARTBEAT_INTERVAL)

		if time.Since(r.LastBeat) > 2*common.INTERNAL_HEARTBEAT_INTERVAL {
			// Close the connection if twice the heartbeat interval has elapsed
			r.Close(4001, "failed to ack heartbeat")
			return
		}
	}
}

// Close is a helper for closing the connection with a close code
func (r *Reflector) Close(code int, msg string) {
	r.Conn.WriteControl(
		websocket.CloseMessage,
		websocket.FormatCloseMessage(code, msg),
		time.Now().Add(time.Second))
	time.Sleep(time.Second)
	r.Conn.Close()
}

func (r *Reflector) Send() {
	for {
		evt := <-r.Events

		r.Conn.SetWriteDeadline(time.Now().Add(time.Second))
		err := r.Conn.WriteMessage(websocket.BinaryMessage, evt)
		if err != nil {
			return
		}
	}
}
