package main

import (
	"crypto/rand"
	"io"

	"github.com/cloudflare/circl/dh/x25519"
	"github.com/cloudflare/circl/pke/kyber/kyber768"
	"github.com/cloudflare/circl/sign/dilithium/mode2"
	"github.com/cloudflare/circl/sign/ed25519"
	"github.com/google/uuid"
)

func main() {
	// alice := GenTx()
	// bob := GenTx()

	// RxFromTx(bob, alice)
	// RxFromTx(alice, bob)

	// b := new(bytes.Buffer)
	// alice.GenerateUpdate(b)
	// bob.Children[0].ReceiveMessage(b.Bytes())

	// b = new(bytes.Buffer)
	// bob.GenerateUpdate(b)
	// alice.Children[0].ReceiveMessage(b.Bytes())

	// b = new(bytes.Buffer)
	// alice.SendMessage([]byte("hello world"), b)
	// out := bob.Children[0].ReceiveMessage(b.Bytes())
	// fmt.Println(string(out))

	// b = new(bytes.Buffer)
	// bob.SendMessage([]byte("hi"), b)
	// out = alice.Children[0].ReceiveMessage(b.Bytes())
	// fmt.Println(string(out))

	genGlobalJS()
	c := make(chan int)
	<-c
}

func GenTx(id uuid.UUID) *TxSession {
	t := &TxSession{UUID: id}

	// Ratchets
	r := Ratchet{}
	var rootRoot ChainKey
	io.ReadFull(rand.Reader, rootRoot[:])
	r.Root = NewRootRatchet(rootRoot)

	var chainRoot ChainKey
	io.ReadFull(rand.Reader, chainRoot[:])
	r.Symmetric = NewSymRatchet(chainRoot)

	t.Ratchets = []*Ratchet{&r}

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

func RxFromTx(local, remote *TxSession) {
	var pub x25519.Key
	x25519.KeyGen(&pub, &remote.CurrentPrivkey)

	var ratchets []*Ratchet
	for _, v := range remote.Ratchets {
		ratchets = append(ratchets, &Ratchet{
			Root:      NewRootRatchet(v.Root.current),
			Symmetric: NewSymRatchet(v.Symmetric.current),
		})
	}

	local.Children = append(local.Children, &RxSession{
		Parent: local,

		UUID:              remote.UUID,
		VerifyingPubkey:   remote.SigningKey.Public().(ed25519.PublicKey),
		VerifyingPubkeyPQ: *remote.SigningKeyPQ.Public().(*mode2.PublicKey),

		Ratchets: ratchets,

		CurrentPubkey:   pub,
		CurrentPubkeyPQ: remote.CurrentPubkeyPQ,
	})
}
