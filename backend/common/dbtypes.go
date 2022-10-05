package common

import "github.com/google/uuid"

type DBGuild struct {
	ID            uuid.UUID `bson:"_id"`
	ReflectorAddr string

	// Map of user ids to list of allowed user ids
	Allowlists map[uuid.UUID][]uuid.UUID
}

type DBReflector struct {
	Addr string
	Load int64 // currently the number of messages relayed every 30 seconds
}

type ByLoad []DBReflector

func (l ByLoad) Len() int           { return len(l) }
func (l ByLoad) Swap(i, j int)      { l[i], l[j] = l[j], l[i] }
func (l ByLoad) Less(i, j int) bool { return l[i].Load < l[j].Load }

type DBUser struct {
	ID    uuid.UUID `bson:"_id"`
	Token []byte
}

type DBMessage struct {
	ID        uuid.UUID `bson:"_id"`
	GuildID   uuid.UUID
	Timestamp int64 // unix millis
	Message   []byte
}
