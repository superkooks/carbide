package main

import (
	"bytes"
	"encoding/binary"
	"io"

	"github.com/cloudflare/circl/dh/x25519"
	"github.com/cloudflare/circl/pke/kyber/kyber768"
	"github.com/cloudflare/circl/sign/dilithium/mode2"
	"github.com/cloudflare/circl/sign/ed25519"
	"github.com/google/uuid"
)

const (
	MSG_TYPE_DATA = iota
	MSG_TYPE_RATCHET_UPDATE
)

// A normal message containing encrypted data
type Data struct {
	SenderID uuid.UUID
	MsgType  byte
	Nonce    [24]byte
	Payload  []byte

	Signature   ECSignature
	SignaturePQ DiLiSignature
}

func (m *Data) Marshal(w io.Writer) {
	w.Write(m.SenderID[:])
	w.Write([]byte{m.MsgType})
	w.Write(m.Nonce[:])
	w.Write(m.Payload)
	w.Write(m.Signature[:])
	w.Write(m.SignaturePQ[:])
}

func (m *Data) Sign(ed ed25519.PrivateKey, dili mode2.PrivateKey) {
	b := new(bytes.Buffer)
	m.Marshal(b)

	dataEnd := b.Len() - ed25519.SignatureSize - mode2.SignatureSize
	msg := b.Bytes()[:dataEnd]

	sig := ed25519.Sign(ed, msg)
	copy(m.Signature[:], sig)

	mode2.SignTo(&dili, msg, m.SignaturePQ[:])
}

func (m *Data) Unmarshal(r io.Reader) {
	io.ReadFull(r, m.SenderID[:])

	b := make([]byte, 1)
	io.ReadFull(r, b)
	m.MsgType = b[0]

	io.ReadFull(r, m.Nonce[:])

	b, _ = io.ReadAll(r)
	m.Payload = b[:len(b)-ed25519.SignatureSize-mode2.SignatureSize]
	newBuf := bytes.NewBuffer(b[len(b)-ed25519.SignatureSize-mode2.SignatureSize:])

	io.ReadFull(newBuf, m.Signature[:])
	io.ReadFull(newBuf, m.SignaturePQ[:])
}

// A message sent for updating the ratchets of other users
type RatchetUpdate struct {
	SenderID    uuid.UUID
	MsgType     byte
	NewPubkey   x25519.Key
	NewPubkeyPQ kyber768.PublicKey

	Updates []UserRatchetUpdate

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
	w.Write([]byte{m.MsgType})
	w.Write(m.NewPubkey[:])

	b := make([]byte, kyber768.PublicKeySize)
	m.NewPubkeyPQ.Pack(b)
	w.Write(b)

	binary.Write(w, binary.BigEndian, int64(len(m.Updates)))
	for _, v := range m.Updates {
		w.Write(v.User[:])
		w.Write(v.DH[:])
		w.Write(v.Kyber[:])
	}

	w.Write(m.Signature[:])
	w.Write(m.SignaturePQ[:])
}

func (m *RatchetUpdate) Sign(ed ed25519.PrivateKey, dili mode2.PrivateKey) {
	b := new(bytes.Buffer)
	m.Marshal(b)

	dataEnd := b.Len() - ed25519.SignatureSize - mode2.SignatureSize
	msg := b.Bytes()[:dataEnd]

	sig := ed25519.Sign(ed, msg)
	copy(m.Signature[:], sig)

	mode2.SignTo(&dili, msg, m.SignaturePQ[:])
}

func (m *RatchetUpdate) Unmarshal(r io.Reader) {
	io.ReadFull(r, m.SenderID[:])
	b := make([]byte, 1)
	io.ReadFull(r, b)
	m.MsgType = b[0]

	io.ReadFull(r, m.NewPubkey[:])

	b = make([]byte, kyber768.PublicKeySize)
	io.ReadFull(r, b)
	m.NewPubkeyPQ.Unpack(b)

	var l int64
	binary.Read(r, binary.BigEndian, &l)
	for i := int64(0); i < l; i++ {
		v := UserRatchetUpdate{}
		io.ReadFull(r, v.User[:])
		io.ReadFull(r, v.DH[:])
		io.ReadFull(r, v.Kyber[:])

		m.Updates = append(m.Updates, v)
	}

	io.ReadFull(r, m.Signature[:])
	io.ReadFull(r, m.SignaturePQ[:])
}
