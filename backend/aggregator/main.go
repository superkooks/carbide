package main

import (
	"carbide/backend/common"
	"context"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const DB_HOST = "127.0.0.1"
const DB_NAME = "carbide"

var db *mongo.Database
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		// TODO: fix this
		return true
	},
}

func main() {
	// Connect to database
	client, err := mongo.NewClient(options.Client().ApplyURI("mongodb://" + DB_HOST).SetRegistry(common.Registry))
	if err != nil {
		panic(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	err = client.Connect(ctx)
	if err != nil {
		panic(err)
	}
	defer client.Disconnect(context.Background())

	err = client.Ping(ctx, nil)
	if err != nil {
		panic(err)
	}

	db = client.Database(DB_NAME)

	// Start http server
	http.HandleFunc("/ws", serveSocket)

	err = http.ListenAndServe(":8080", nil)
	if err != nil {
		panic(err)
	}
}

func serveSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		panic(err)
	}

	c := NewClient(conn)
	go c.Listen()
	go c.Send()
	go c.Heartbeat()
}
