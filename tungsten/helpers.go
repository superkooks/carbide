package main

import (
	"bytes"
	"io"

	"github.com/google/uuid"
)

type EvtType byte

const (
	DATA = iota
	SUB_GUILDS
	ADD_USER
)

func MarshalData(guild string, msg []byte) []byte {
	b := new(bytes.Buffer)

	b.Write([]byte{0x00})

	parsed, err := uuid.Parse(guild)
	if err != nil {
		panic(err)
	}
	b.Write(parsed[:])

	b.Write(msg)

	return b.Bytes()
}

func UnmarshalData(event []byte) (string, []byte) {
	b := bytes.NewBuffer(event)

	t, _ := b.ReadByte()
	if t != DATA {
		panic("unmarshal data: event is not type data")
	}

	var guild uuid.UUID
	b.Read(guild[:])

	msg, _ := io.ReadAll(b)

	return guild.String(), msg
}

func MarshalSubGuilds(uuids []string) []byte {
	b := new(bytes.Buffer)

	b.Write([]byte{0x01})

	for _, v := range uuids {
		parsed, err := uuid.Parse(v)
		if err != nil {
			panic(err)
		}

		b.Write(parsed[:])
	}

	return b.Bytes()
}
