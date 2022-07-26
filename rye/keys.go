package main

import (
	"github.com/cloudflare/circl/pke/kyber/kyber768"
	"golang.org/x/crypto/nacl/secretbox"
)

type ChainKey [32]byte
type MessageKey [32]byte

// Keys encapsulated by a method
type DHKey [32]byte
type KyberKey [32]byte

// The output of encapsulation by a method
type DHKeyCiphertext [32 + 24 + secretbox.Overhead]byte
type KyberKeyCiphertext [kyber768.CiphertextSize]byte
