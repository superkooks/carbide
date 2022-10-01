package main

import (
	"bytes"
	"encoding/binary"
	"io"

	"github.com/google/uuid"
)

const (
	HEARTBEAT = iota
	HEARTBEAT_ACK
	ERROR
	DATA
	DATA_ACK
	REGISTER
	AUTHENTICATE
	SUB_GUILDS
	ADD_USERS
	REMOVE_USERS
)

func MarshalData(guild string, msg []byte) []byte {
	b := new(bytes.Buffer)

	b.Write([]byte{DATA})

	parsed, err := uuid.Parse(guild)
	if err != nil {
		panic(err)
	}
	b.Write(parsed[:])

	b.Write(msg)

	return b.Bytes()
}

func UnmarshalData(event []byte) (guild string, evt string, ts int64, msg []byte) {
	b := bytes.NewBuffer(event)

	t, _ := b.ReadByte()
	if t != DATA {
		panic("unmarshal data: event is not type DATA")
	}

	var guildID uuid.UUID
	b.Read(guildID[:])

	var evtID uuid.UUID
	b.Read(evtID[:])

	binary.Read(b, binary.BigEndian, ts)

	msg, _ = io.ReadAll(b)
	return
}

func UnmarshalDataAck(event []byte) (guild string, evt string, ts int64) {
	b := bytes.NewBuffer(event)

	t, _ := b.ReadByte()
	if t != DATA_ACK {
		panic("unmarshal data: event is not type DATA_ACK")
	}

	var guildID uuid.UUID
	b.Read(guildID[:])

	var evtID uuid.UUID
	b.Read(evtID[:])

	binary.Read(b, binary.BigEndian, ts)

	return
}

func UnmarshalRegister(event []byte) (user string, token []byte) {
	b := bytes.NewBuffer(event)

	t, _ := b.ReadByte()
	if t != REGISTER {
		panic("unmarshal data: event is not type REGISTER")
	}

	var userID uuid.UUID
	b.Read(userID[:])

	token, _ = io.ReadAll(b)

	return
}

func MarshalSubGuilds(uuids []string) []byte {
	b := new(bytes.Buffer)

	b.Write([]byte{SUB_GUILDS})

	for _, v := range uuids {
		parsed, err := uuid.Parse(v)
		if err != nil {
			panic(err)
		}

		b.Write(parsed[:])
	}

	return b.Bytes()
}

func MarshalAddUsers(guild string, uuids []string) []byte {
	b := new(bytes.Buffer)

	b.Write([]byte{ADD_USERS})

	parsed, err := uuid.Parse(guild)
	if err != nil {
		panic(err)
	}
	b.Write(parsed[:])

	for _, v := range uuids {
		parsed, err := uuid.Parse(v)
		if err != nil {
			panic(err)
		}

		b.Write(parsed[:])
	}

	return b.Bytes()
}

func MarshalRemoveUsers(guild string, uuids []string) []byte {
	b := new(bytes.Buffer)

	b.Write([]byte{REMOVE_USERS})

	parsed, err := uuid.Parse(guild)
	if err != nil {
		panic(err)
	}
	b.Write(parsed[:])

	for _, v := range uuids {
		parsed, err := uuid.Parse(v)
		if err != nil {
			panic(err)
		}

		b.Write(parsed[:])
	}

	return b.Bytes()
}
