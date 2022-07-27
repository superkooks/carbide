package main

import (
	"crypto/sha256"
	"io"

	"github.com/cloudflare/circl/dh/x25519"
	"github.com/cloudflare/circl/pke/kyber/kyber768"
	"github.com/cloudflare/circl/sign/dilithium/mode2"
	"github.com/cloudflare/circl/sign/ed25519"
	"golang.org/x/crypto/hkdf"
	"golang.org/x/crypto/nacl/secretbox"
)

type RxSession struct {
	Parent *TxSession

	VerifyingPubkey   ed25519.PublicKey
	VerifyingPubkeyPQ mode2.PublicKey

	Symmetric *SymRatchet
	Root      *RootRatchet
}

func (r *RxSession) DecryptMessage(msg []byte) {
	// Verify both signatures
	dataEnd := len(msg) - 1 - ed25519.SignatureSize - mode2.SignatureSize

	ok := ed25519.Verify(r.VerifyingPubkey, msg[:dataEnd], msg[dataEnd:dataEnd+ed25519.SignatureSize])
	if !ok {
		panic("failed to verify ed25519 signature")
	}

	ok = mode2.Verify(&r.VerifyingPubkeyPQ, msg[:dataEnd], msg[len(msg)-1-mode2.SignatureSize:len(msg)-1])
	if !ok {
		panic("failed to verify dilithium mode2 signature")
	}
}

func (r *RxSession) UpdateSymmetric(newPub x25519.Key, newPubPQ kyber768.PublicKey, dhIn DHKeyCiphertext, kyberIn KyberKeyCiphertext) {
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
	secretbox.Open(encap[:], dhIn[:], &nonce, &derived)

	var encapPQ KyberKey
	r.Parent.CurrentPrivkeyPQ.DecryptTo(encapPQ[:], kyberIn[:])

	// Advance our root ratchet
	chain := r.Root.Advance(encap, encapPQ)

	// Generate new symmetric ratchet
	r.Symmetric = NewSymRatchet(chain)
}
