package common

import (
	"time"

	"github.com/google/uuid"
)

const INTERNAL_HEARTBEAT_INTERVAL = 5 * time.Second
const CLIENT_HEARTBEAT_INTERVAL = 20 * time.Second

type EventType byte

const (
	EVT_HEARTBEAT EventType = iota
	EVT_HEARTBEAT_ACK
	EVT_ERROR
	EVT_DATA
	EVT_REGISTER
	EVT_AUTHENTICATE
	EVT_SUB_GUILDS
	EVT_ADD_USERS
	EVT_REMOVE_USERS
)

type Event struct {
	Type EventType `msgpack:"type"`
	Evt  []byte    `msgpack:"evt"`
}

// Websocket close codes
// 4000 - failed to authenticate
// 4001 - failed to ack heartbeat

// When a client receives a heartbeat, it should immediately respond with a Heartbeat Ack.
// Sent by the backend only.
// type Heartbeat struct {}

// Sent by clients only.
// type HearbeatAck struct {}

// An error which does not result in the connection being closed.
// For errors that do, see the close codes
// Sent by the backend only.
type EvtError struct {
	Code int `msgpack:"code"`
}

// Clients can know their message was distributed without having the decrypt their own
// message using the EvtID field. It should be set to a random value, and stored alongside
// the mutation until the same event id was processed, indicating the mutation should be applied.
// Sent by clients and the backend.
type EvtData struct {
	GuildID   uuid.UUID `msgpack:"guildId"`
	EvtID     uuid.UUID `msgpack:"evtId"`     // set by client
	Timestamp int64     `msgpack:"timestamp"` // unix millis, left 0 by client
	Message   []byte    `msgpack:"message"`
}

// Generate new user uuid and corresponding token.
// The backend will return a filled version of the event.
// The connection will be authenticated once the event is returned
// Sent by clients and the backend.
type EvtRegister struct {
	UserID uuid.UUID `msgpack:"userId"`
	Token  []byte    `msgpack:"token"`
}

// Sent by clients only.
type EvtAuthenticate struct {
	Token []byte `msgpack:"token"`
}

// Subscribe the client to the specified guilds.
// Sent by clients only.
type EvtSubGuilds struct {
	GuildIDs []uuid.UUID `msgpack:"guildIds"`
}

// Add the specified users to the user's guild allowlist.
// Sent by clients only.
type EvtAddUsers struct {
	GuildID uuid.UUID   `msgpack:"guildId"`
	UserIDs []uuid.UUID `msgpack:"userIds"`
}

// Remove the specified users to the user's guild allowlist.
// Sent by clients only.
type EvtRemoveUsers struct {
	GuildID uuid.UUID   `msgpack:"guildId"`
	UserIDs []uuid.UUID `msgpack:"userIds"`
}
