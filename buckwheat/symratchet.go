package main

import (
	"crypto/hmac"
	"crypto/sha256"
)

var RATCHET_HMAC_MSG = []byte{0x01}
var RATCHET_HMAC_CHAIN = []byte{0x02}

type SymRatchet struct {
	current ChainKey
}

func NewSymRatchet(root ChainKey) *SymRatchet {
	return &SymRatchet{
		current: root,
	}
}

func (r *SymRatchet) Advance() MessageKey {
	h := hmac.New(sha256.New, r.current[:])
	h.Write(RATCHET_HMAC_CHAIN)
	next := h.Sum(nil)
	h.Reset()

	h.Write(RATCHET_HMAC_MSG)
	msgKey := h.Sum(nil)

	copy(r.current[:], next)

	var out MessageKey
	copy(out[:], msgKey)
	return out
}
