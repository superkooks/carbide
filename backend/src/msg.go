package main

import (
	"io"

	"github.com/google/uuid"
)

const (
	EVT_TYPE_DATA = iota
	EVT_TYPE_SUBSCRIBE_GUILDS
	EVT_TYPE_ADD_USER
)

// A normal message
type Data struct {
	EvtType byte
	Guild   uuid.UUID
	Message []byte
}

func (d *Data) Marshal(w io.Writer) {
	w.Write([]byte{d.EvtType})
	w.Write(d.Guild[:])
	w.Write(d.Message)
}

func (d *Data) Unmarshal(r io.Reader) {
	b := make([]byte, 1)
	r.Read(b)
	d.EvtType = b[0]

	io.ReadFull(r, d.Guild[:])
	d.Message, _ = io.ReadAll(r)
}
