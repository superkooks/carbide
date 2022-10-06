package main

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"errors"
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
	m := Data{SenderID: t.UUID, MsgType: MSG_TYPE_DATA}
	io.ReadFull(rand.Reader, m.Nonce[:])

	key := t.Symmetric.Advance()
	m.Payload = secretbox.Seal(nil, msg, &m.Nonce, (*[32]byte)(&key))

	m.Sign(t.SigningKey, t.SigningKeyPQ)
	m.Marshal(w)
}

func (t *TxSession) ReceiveMessage(msg []byte) ([]byte, error) {
	var u uuid.UUID
	copy(u[:], msg)

	for _, v := range t.Children {
		if v.UUID == u {
			return v.ReceiveMessage(msg), nil
		}
	}

	return []byte{}, errors.New("couldn't find rx for message")
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

func (t *TxSession) Export(w io.Writer) {
	w.Write(t.UUID[:])
	w.Write(t.SigningKey)
	w.Write(t.SigningKeyPQ.Bytes())
	w.Write(t.Symmetric.current[:])
	w.Write(t.Root.current[:])
	w.Write(t.CurrentPrivkey[:])

	privPQ := make([]byte, kyber768.PrivateKeySize)
	t.CurrentPrivkeyPQ.Pack(privPQ)
	w.Write(privPQ)

	pubPQ := make([]byte, kyber768.PublicKeySize)
	t.CurrentPubkeyPQ.Pack(pubPQ)
	w.Write(pubPQ)

	binary.Write(w, binary.BigEndian, int64(len(t.Children)))
	for _, v := range t.Children {
		v.Export(w)
	}
}

func ImportTx(r io.Reader) *TxSession {
	t := new(TxSession)

	r.Read(t.UUID[:])

	t.SigningKey = make(ed25519.PrivateKey, ed25519.PrivateKeySize)
	r.Read(t.SigningKey)

	var sigPQ [mode2.PrivateKeySize]byte
	r.Read(sigPQ[:])
	t.SigningKeyPQ.Unpack(&sigPQ)

	var symChain ChainKey
	r.Read(symChain[:])
	t.Symmetric = NewSymRatchet(symChain)

	var rootChain ChainKey
	r.Read(rootChain[:])
	t.Root = NewRootRatchet(rootChain)

	r.Read(t.CurrentPrivkey[:])

	privPQ := make([]byte, kyber768.PrivateKeySize)
	r.Read(privPQ)
	t.CurrentPrivkeyPQ.Unpack(privPQ)

	pubPQ := make([]byte, kyber768.PublicKeySize)
	r.Read(pubPQ)
	t.CurrentPubkeyPQ.Unpack(pubPQ)

	var childrenCount int64
	binary.Read(r, binary.BigEndian, &childrenCount)
	for i := 0; i < int(childrenCount); i++ {
		rx := ImportRx(r)
		rx.Parent = t
		t.Children = append(t.Children, rx)
	}

	return t
}
