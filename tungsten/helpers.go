package main

import (
	"bytes"

	"github.com/google/uuid"
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
