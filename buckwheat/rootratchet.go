package main

import (
	"crypto/hmac"
	"crypto/sha256"
)

type RootRatchet struct {
	current ChainKey
}

func NewRootRatchet(root ChainKey) *RootRatchet {
	return &RootRatchet{
		current: root,
	}
}

func (r *RootRatchet) Advance(dh DHKey, kyber KyberKey) ChainKey {
	h := hmac.New(sha256.New, append(append(r.current[:], dh[:]...), kyber[:]...))
	h.Write(RATCHET_HMAC_CHAIN)
	next := h.Sum(nil)
	h.Reset()

	h.Write(RATCHET_HMAC_MSG)
	msgKey := h.Sum(nil)

	copy(r.current[:], next)

	var out ChainKey
	copy(out[:], msgKey)
	return out
}
