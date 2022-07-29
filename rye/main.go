package main

import (
	"bytes"
	"crypto/rand"
	"fmt"
	"io"

	"github.com/cloudflare/circl/dh/x25519"
	"github.com/cloudflare/circl/pke/kyber/kyber768"
	"github.com/cloudflare/circl/sign/dilithium/mode2"
	"github.com/cloudflare/circl/sign/ed25519"
	"github.com/google/uuid"
)

func main() {
	alice := GenTX()
	bob := GenTX()

	RXFromTX(bob, alice)
	RXFromTX(alice, bob)

	b := new(bytes.Buffer)
	alice.GenerateUpdate(b)
	bob.Children[0].ReceiveMessage(b.Bytes())

	b = new(bytes.Buffer)
	bob.GenerateUpdate(b)
	alice.Children[0].ReceiveMessage(b.Bytes())

	b = new(bytes.Buffer)
	alice.SendMessage([]byte("hello world"), b)
	out := bob.Children[0].ReceiveMessage(b.Bytes())
	fmt.Println(string(out))
}

func GenTX() *TxSession {
	t := &TxSession{UUID: uuid.New()}

	// Ratchets
	var rootRoot ChainKey
	io.ReadFull(rand.Reader, rootRoot[:])
	t.Root = NewRootRatchet(rootRoot)

	var chainRoot ChainKey
	io.ReadFull(rand.Reader, chainRoot[:])
	t.Symmetric = NewSymRatchet(chainRoot)

	// Signing keys
	_, t.SigningKey, _ = ed25519.GenerateKey(rand.Reader)

	_, priv, _ := mode2.GenerateKey(rand.Reader)
	t.SigningKeyPQ = *priv

	// Key encap keys
	io.ReadFull(rand.Reader, t.CurrentPrivkey[:])

	public, private, err := kyber768.GenerateKey(rand.Reader)
	if err != nil {
		panic(err)
	}
	t.CurrentPrivkeyPQ = *private
	t.CurrentPubkeyPQ = *public

	return t
}

func RXFromTX(local, remote *TxSession) {
	var pub x25519.Key
	x25519.KeyGen(&pub, &remote.CurrentPrivkey)

	local.Children = append(local.Children, &RxSession{
		Parent: local,

		UUID:              remote.UUID,
		VerifyingPubkey:   remote.SigningKey.Public().(ed25519.PublicKey),
		VerifyingPubkeyPQ: *remote.SigningKeyPQ.Public().(*mode2.PublicKey),

		Root:      NewRootRatchet(remote.Root.current),
		Symmetric: NewSymRatchet(remote.Symmetric.current),

		CurrentPubkey:   pub,
		CurrentPubkeyPQ: remote.CurrentPubkeyPQ,
	})
}
