package main

import "github.com/google/uuid"

var guilds = make(map[uuid.UUID]*Guild)

type Guild struct {
	Clients []*Client
}
