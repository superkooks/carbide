package main

import (
	"crypto/rand"
	"crypto/sha256"
	"io"

	"github.com/cloudflare/circl/dh/x25519"
	"github.com/cloudflare/circl/pke/kyber/kyber768"
	"github.com/cloudflare/circl/sign/dilithium/mode2"
	"github.com/cloudflare/circl/sign/ed25519"
	"github.com/google/uuid"
	"golang.org/x/crypto/hkdf"
	"golang.org/x/crypto/nacl/secretbox"
)

var DH_HKDF_INFO = []byte("dh_hkdf")

type TxSession struct {
	UUID         uuid.UUID
	SigningKey   ed25519.PrivateKey
	SigningKeyPQ mode2.PrivateKey

	Symmetric *SymRatchet
	Root      *RootRatchet

	CurrentPrivkey   x25519.Key
	CurrentPrivkeyPQ kyber768.PrivateKey
	CurrentPubkeyPQ  kyber768.PublicKey

	Children []*RxSession
}

func (t *TxSession) SendMessage(msg []byte, w io.Writer) {
	m := Message{SenderID: t.UUID, MsgType: MSG_TYPE_NORMAL}
	io.ReadFull(rand.Reader, m.Nonce[:])

	key := t.Symmetric.Advance()
	m.Payload = secretbox.Seal(nil, msg, &m.Nonce, (*[32]byte)(&key))

	m.Sign(t.SigningKey, t.SigningKeyPQ)
	m.Marshal(w)
}

func (t *TxSession) GenerateUpdate(out io.Writer) {
	// Generate new keypairs
	io.ReadFull(rand.Reader, t.CurrentPrivkey[:])

	pub, priv, err := kyber768.GenerateKey(rand.Reader)
	if err != nil {
		panic(err)
	}
	t.CurrentPrivkeyPQ = *priv
	t.CurrentPubkeyPQ = *pub

	// Generate random keys
	var encap DHKey
	var encapPQ KyberKey
	io.ReadFull(rand.Reader, encap[:])
	io.ReadFull(rand.Reader, encapPQ[:])

	// Start message
	var newPub x25519.Key
	x25519.KeyGen(&newPub, &t.CurrentPrivkey)

	u := &RatchetUpdate{
		SenderID:    t.UUID,
		MsgType:     MSG_TYPE_RATCHET_UPDATE,
		NewPubkey:   newPub,
		NewPubkeyPQ: *pub,
	}

	// Encrypt keys to each other user
	for _, v := range t.Children {
		// Find DH shared secret and derive symmetric key
		var shared x25519.Key
		x25519.Shared(&shared, &t.CurrentPrivkey, &v.CurrentPubkey)

		var derived [32]byte
		keyReader := hkdf.New(sha256.New, shared[:], nil, DH_HKDF_INFO)
		_, err = io.ReadFull(keyReader, derived[:])
		if err != nil {
			panic(err)
		}

		// Encapsulate them (nonce is prepended to ciphertext)
		var outDH DHKeyCiphertext
		var nonce [24]byte
		io.ReadFull(rand.Reader, nonce[:])
		copy(outDH[:], nonce[:])
		ciphertext := secretbox.Seal(nil, encap[:], &nonce, &derived)
		copy(outDH[24:], ciphertext)

		var outKyber KyberKeyCiphertext
		var seed [kyber768.EncryptionSeedSize]byte
		io.ReadFull(rand.Reader, seed[:])
		v.CurrentPubkeyPQ.EncryptTo(outKyber[:], encapPQ[:], seed[:])

		// Add to message
		u.Updates = append(u.Updates, UserRatchetUpdate{
			User:  v.UUID,
			DH:    outDH,
			Kyber: outKyber,
		})
	}

	// Advance our root ratchet
	chain := t.Root.Advance(encap, encapPQ)

	// Generate new symmetric ratchet
	t.Symmetric = NewSymRatchet(chain)

	u.Sign(t.SigningKey, t.SigningKeyPQ)
	u.Marshal(out)
}
