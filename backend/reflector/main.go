package main

import (
	"carbide/backend/common"
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/websocket"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

//lint:file-ignore SA1012 Mongo permits it

const DB_HOST = "127.0.0.1"
const DB_NAME = "carbide"

const REFLECTOR_PORT = "7460"

// messagesRelayed is the indicator of load aggregators use for allocating reflectores
var messagesRelayed int64
var db *mongo.Database
var upgrader = websocket.Upgrader{}
var externalAddr string

func main() {
	// Find external ip
	externalAddr = getOutboundIP().String()
	if _, exists := os.LookupEnv("CARBIDE_REFLECTOR_USE_LOCALHOST"); exists {
		externalAddr = "localhost"
	}
	externalAddr += ":" + REFLECTOR_PORT
	fmt.Println("using external addr", externalAddr)

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
	go func() {
		err = http.ListenAndServe(":"+REFLECTOR_PORT, nil)
		if err != nil {
			panic(err)
		}
	}()

	// Start advertising ourselves
	updateLoad()
}

func serveSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		panic(err)
	}

	a := NewAggregator(conn)
	go a.Listen()
	go a.Send()
	go a.Heartbeat()
}

func updateLoad() {
	for {
		_, err := db.Collection("reflectors").ReplaceOne(nil, bson.D{{Key: "addr", Value: externalAddr}}, common.DBReflector{
			Addr: externalAddr,
			Load: messagesRelayed,
		}, options.Replace().SetUpsert(true))
		if err != nil {
			panic(err)
		}

		messagesRelayed = 0

		time.Sleep(time.Second * 30)
	}
}

// Get preferred outbound ip of this machine
// https://stackoverflow.com/questions/23558425/how-do-i-get-the-local-ip-address-in-go#23558495
func getOutboundIP() net.IP {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)

	return localAddr.IP
}
