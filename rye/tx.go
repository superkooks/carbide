package main

import (
	"crypto/rand"
	"crypto/sha256"
	"io"

	"github.com/cloudflare/circl/dh/x25519"
	"github.com/cloudflare/circl/pke/kyber/kyber768"
	"github.com/cloudflare/circl/sign/dilithium/mode2"
	"github.com/cloudflare/circl/sign/ed25519"
	"golang.org/x/crypto/hkdf"
	"golang.org/x/crypto/nacl/secretbox"
)

var DH_HKDF_INFO = []byte("dh_hkdf")

type TxSession struct {
	SigningKey   ed25519.PrivateKey
	SigningKeyPQ mode2.PrivateKey

	Symmetric *SymRatchet
	Root      *RootRatchet

	CurrentPrivkey   x25519.Key
	CurrentPrivkeyPQ kyber768.PrivateKey
	OthersPubkey     []x25519.Key
}

func (t *TxSession) UpdateSymmetric(target x25519.Key, targetPQ kyber768.PublicKey) (dh DHKeyCiphertext, kyber KyberKeyCiphertext) {
	// Generate new keypairs
	io.ReadFull(rand.Reader, t.CurrentPrivkey[:])

	_, priv, err := kyber768.GenerateKey(rand.Reader)
	if err != nil {
		panic(err)
	}
	t.CurrentPrivkeyPQ = *priv

	// Find DH shared secret and derive symmetric key
	var shared x25519.Key
	x25519.Shared(&shared, &t.CurrentPrivkey, &target)

	var derived [32]byte
	keyReader := hkdf.New(sha256.New, shared[:], nil, DH_HKDF_INFO)
	_, err = io.ReadFull(keyReader, derived[:])
	if err != nil {
		panic(err)
	}

	// Generate random keys
	var encap DHKey
	var encapPQ KyberKey
	io.ReadFull(rand.Reader, encap[:])
	io.ReadFull(rand.Reader, encapPQ[:])

	// Encapsulate them (nonce is prepended to ciphertext)
	var outDH DHKeyCiphertext
	var nonce [24]byte
	io.ReadFull(rand.Reader, nonce[:])
	copy(outDH[:], nonce[:])
	secretbox.Seal(outDH[24:], encap[:], &nonce, &derived)

	var outKyber KyberKeyCiphertext
	var seed [kyber768.EncryptionSeedSize]byte
	io.ReadFull(rand.Reader, seed[:])
	targetPQ.EncryptTo(outKyber[:], encapPQ[:], seed[:])

	// Advance our root ratchet
	chain := t.Root.Advance(encap, encapPQ)

	// Generate new symmetric ratchet
	t.Symmetric = NewSymRatchet(chain)

	return outDH, outKyber
}
