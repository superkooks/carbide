package main

import (
	"bytes"
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

type RxSession struct {
	Parent *TxSession

	UUID              uuid.UUID
	VerifyingPubkey   ed25519.PublicKey
	VerifyingPubkeyPQ mode2.PublicKey

	Symmetric *SymRatchet
	Root      *RootRatchet

	CurrentPubkey   x25519.Key
	CurrentPubkeyPQ kyber768.PublicKey
}

func (r *RxSession) ReceiveMessage(msg []byte) []byte {
	// Verify both signatures
	m := new(Data)
	m.Unmarshal(bytes.NewBuffer(msg))
	dataEnd := len(msg) - ed25519.SignatureSize - mode2.SignatureSize

	ok := ed25519.Verify(r.VerifyingPubkey, msg[:dataEnd], m.Signature[:])
	if !ok {
		panic("failed to verify ed25519 signature")
	}

	ok = mode2.Verify(&r.VerifyingPubkeyPQ, msg[:dataEnd], m.SignaturePQ[:])
	if !ok {
		panic("failed to verify dilithium mode2 signature")
	}

	// Switch on message type
	switch m.MsgType {
	case MSG_TYPE_DATA:
		key := r.Symmetric.Advance()

		plain, ok := secretbox.Open(nil, m.Payload, &m.Nonce, (*[32]byte)(&key))
		if !ok {
			panic("failed to verify mac of payload")
		}

		return plain

	case MSG_TYPE_RATCHET_UPDATE:
		u := new(RatchetUpdate)
		u.Unmarshal(bytes.NewBuffer(msg))

		for _, v := range u.Updates {
			if v.User == r.Parent.UUID {
				r.UpdateSymmetric(u.NewPubkey, u.NewPubkeyPQ, v.DH, v.Kyber)
				break
			}
		}
	}

	return []byte{}
}

func (r *RxSession) UpdateSymmetric(newPub x25519.Key, newPubPQ kyber768.PublicKey, dhIn DHKeyCiphertext, kyberIn KyberKeyCiphertext) {
	// Update current pubkeys
	r.CurrentPubkey = newPub
	r.CurrentPubkeyPQ = newPubPQ

	// Find DH shared secret and derive symmetric key
	var shared x25519.Key
	x25519.Shared(&shared, &r.Parent.CurrentPrivkey, &newPub)

	var derived [32]byte
	keyReader := hkdf.New(sha256.New, shared[:], nil, DH_HKDF_INFO)
	_, err := io.ReadFull(keyReader, derived[:])
	if err != nil {
		panic(err)
	}

	// Unencapsulate them
	var encap DHKey
	var nonce [24]byte
	copy(nonce[:], dhIn[:])
	plain, ok := secretbox.Open(nil, dhIn[24:], &nonce, &derived)
	if !ok {
		panic("failed to verify mac of encapsulated key")
	}
	copy(encap[:], plain)

	var encapPQ KyberKey
	r.Parent.CurrentPrivkeyPQ.DecryptTo(encapPQ[:], kyberIn[:])

	// Advance our root ratchet
	chain := r.Root.Advance(encap, encapPQ)

	// Generate new symmetric ratchet
	r.Symmetric = NewSymRatchet(chain)
}

func (r *RxSession) Export(w io.Writer) {
	w.Write(r.UUID[:])
	w.Write(r.VerifyingPubkey)
	w.Write(r.VerifyingPubkeyPQ.Bytes())
	w.Write(r.Symmetric.current[:])
	w.Write(r.Root.current[:])
	w.Write(r.CurrentPubkey[:])

	curPubPQ := make([]byte, kyber768.PublicKeySize)
	r.CurrentPubkeyPQ.Pack(curPubPQ)
	w.Write(curPubPQ)
}

func ImportRx(i io.Reader) *RxSession {
	r := new(RxSession)

	i.Read(r.UUID[:])

	r.VerifyingPubkey = make(ed25519.PublicKey, ed25519.PublicKeySize)
	i.Read(r.VerifyingPubkey)

	var verPubPQ [mode2.PublicKeySize]byte
	i.Read(verPubPQ[:])
	r.VerifyingPubkeyPQ.Unpack(&verPubPQ)

	var symChain ChainKey
	i.Read(symChain[:])
	r.Symmetric = NewSymRatchet(symChain)

	var rootChain ChainKey
	i.Read(rootChain[:])
	r.Root = NewRootRatchet(rootChain)

	i.Read(r.CurrentPubkey[:])

	curPubPQ := make([]byte, kyber768.PublicKeySize)
	i.Read(curPubPQ)
	r.CurrentPubkeyPQ.Unpack(curPubPQ)

	return r
}
