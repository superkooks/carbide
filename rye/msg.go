package main

import (
	"encoding/binary"
	"io"

	"github.com/cloudflare/circl/dh/x25519"
	"github.com/cloudflare/circl/pke/kyber/kyber768"
	"github.com/google/uuid"
)

// A normal message containing encrypted data
type Message struct {
	SenderID uuid.UUID
	Nonce    []byte
	Payload  []byte

	Signature   ECSignature
	SignaturePQ DiLiSignature
}

func (m *Message) Marshal(w io.Writer) {
	w.Write(m.SenderID[:])
	w.Write(m.Nonce)
	w.Write(m.Payload)
	w.Write(m.Signature[:])
	w.Write(m.SignaturePQ[:])
}

func (m *Message) Unmarshal(r io.Reader) {
	io.ReadFull(r, m.SenderID[:])
	io.ReadFull(r, m.Nonce)
	io.ReadFull(r, m.Payload)
	io.ReadFull(r, m.Signature[:])
	io.ReadFull(r, m.SignaturePQ[:])
}

// A message sent for updating the ratchets of other users
type RatchetUpdate struct {
	SenderID    uuid.UUID
	NewPubkey   x25519.Key
	NewPubkeyPQ kyber768.PublicKey

	UpdatesLen int64
	Updates    []UserRatchetUpdate

	Signature   ECSignature
	SignaturePQ DiLiSignature
}

// Part of RatchetUpdate. Addressed per user.
type UserRatchetUpdate struct {
	User  uuid.UUID
	DH    DHKeyCiphertext
	Kyber KyberKeyCiphertext
}

func (m *RatchetUpdate) Marshal(w io.Writer) {
	w.Write(m.SenderID[:])
	w.Write(m.NewPubkey[:])

	b := make([]byte, kyber768.PublicKeySize)
	m.NewPubkeyPQ.Pack(b)
	w.Write(b)

	binary.Write(w, binary.BigEndian, m.UpdatesLen)
	for _, v := range m.Updates {
		w.Write(v.User[:])
		w.Write(v.DH[:])
		w.Write(v.Kyber[:])
	}

	w.Write(m.Signature[:])
	w.Write(m.SignaturePQ[:])
}

func (m *RatchetUpdate) Unmarshal(r io.Reader) {
	io.ReadFull(r, m.SenderID[:])
	io.ReadFull(r, m.NewPubkey[:])

	b := make([]byte, kyber768.PublicKeySize)
	io.ReadFull(r, b)
	m.NewPubkeyPQ.Unpack(b)

	binary.Read(r, binary.BigEndian, m.UpdatesLen)
	for i := int64(0); i < m.UpdatesLen; i++ {
		v := UserRatchetUpdate{}
		io.ReadFull(r, v.User[:])
		io.ReadFull(r, v.DH[:])
		io.ReadFull(r, v.Kyber[:])

		m.Updates = append(m.Updates, v)
	}

	io.ReadFull(r, m.Signature[:])
	io.ReadFull(r, m.SignaturePQ[:])
}
