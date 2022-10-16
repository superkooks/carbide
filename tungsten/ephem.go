package main

import (
	"bytes"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"io"
	"math/big"
	"strings"

	"github.com/cloudflare/circl/dh/x25519"
	"github.com/cloudflare/circl/pke/kyber/kyber768"
	"golang.org/x/crypto/argon2"
	"golang.org/x/crypto/hkdf"
	"golang.org/x/crypto/nacl/secretbox"

	_ "embed"
)

var DH_HKDF_EPHEM = []byte("dh_hkdf_ephem")
var EPHEM_HMAC = []byte{0x03}
var EPHEM_FINGERPRINT_SALT = []byte("ephem_fingerprint_salt")

// The private part of an ephem keypair
type EphemPriv struct {
	Privkey   x25519.Key
	PrivkeyPQ kyber768.PrivateKey
	PubkeyPQ  kyber768.PublicKey
}

func (e *EphemPriv) Marshal(w io.Writer) {
	w.Write(e.Privkey[:])

	pq := make([]byte, kyber768.PrivateKeySize)
	e.PrivkeyPQ.Pack(pq)
	w.Write(pq)

	pq = make([]byte, kyber768.PublicKeySize)
	e.PubkeyPQ.Pack(pq)
	w.Write(pq)
}

func (e *EphemPriv) Unmarshal(r io.Reader) {
	r.Read(e.Privkey[:])

	pq := make([]byte, kyber768.PrivateKeySize)
	io.ReadFull(r, pq)
	e.PrivkeyPQ.Unpack(pq)

	pq = make([]byte, kyber768.PublicKeySize)
	io.ReadFull(r, pq)
	e.PubkeyPQ.Unpack(pq)
}

// The public part of an ephem keypair
type EphemPub struct {
	Pubkey   x25519.Key
	PubkeyPQ kyber768.PublicKey
}

func (e *EphemPub) Marshal(w io.Writer) {
	w.Write(e.Pubkey[:])

	pq := make([]byte, kyber768.PublicKeySize)
	e.PubkeyPQ.Pack(pq)
	w.Write(pq)
}

func (e *EphemPub) Unmarshal(r io.Reader) {
	r.Read(e.Pubkey[:])

	pq := make([]byte, kyber768.PublicKeySize)
	io.ReadFull(r, pq)
	e.PubkeyPQ.Unpack(pq)
}

// GenEphem generates an ephem keypair
func GenEphem() (*EphemPriv, *EphemPub) {
	priv := new(EphemPriv)
	pub := new(EphemPub)

	// Generate x25519 keys
	io.ReadFull(rand.Reader, priv.Privkey[:])
	x25519.KeyGen(&pub.Pubkey, &priv.Privkey)

	// Generate kyber keys
	pubpq, privpq, err := kyber768.GenerateKey(rand.Reader)
	if err != nil {
		panic(err)
	}
	pub.PubkeyPQ = *pubpq
	priv.PrivkeyPQ = *privpq

	return priv, pub
}

func GenerateSharedSecret(local *EphemPriv, remote *EphemPub) (ciphertext []byte, secret [32]byte) {
	// Generate sub shared-secrets
	var encap DHKey
	var encapPQ KyberKey
	io.ReadFull(rand.Reader, encap[:])
	io.ReadFull(rand.Reader, encapPQ[:])

	// Find DH shared secret
	var dhShared x25519.Key
	x25519.Shared(&dhShared, &local.Privkey, &remote.Pubkey)

	var derived [32]byte
	keyReader := hkdf.New(sha256.New, append(dhShared[:]), nil, DH_HKDF_EPHEM)
	_, err := io.ReadFull(keyReader, derived[:])
	if err != nil {
		panic(err)
	}

	// Encapsulate them (nonce is prepended to ciphertext)
	var outDH DHKeyCiphertext
	var nonce [24]byte
	io.ReadFull(rand.Reader, nonce[:])
	copy(outDH[:], nonce[:])
	dhCtext := secretbox.Seal(nil, encap[:], &nonce, &derived)
	copy(outDH[24:], dhCtext)

	var outKyber KyberKeyCiphertext
	local.PubkeyPQ.EncryptTo(outKyber[:], encapPQ[:], nil)

	// Derive shared secret
	h := hmac.New(sha256.New, append(outDH[:], outKyber[:]...))
	h.Write(EPHEM_HMAC)
	var shared MessageKey
	copy(shared[:], h.Sum(nil))

	return append(outDH[:], outKyber[:]...), shared
}

func ReceiveSharedSecret(local *EphemPriv, remote *EphemPub, ciphertext []byte) [32]byte {
	// Find DH shared secret
	var dhShared x25519.Key
	x25519.Shared(&dhShared, &local.Privkey, &remote.Pubkey)

	var derived [32]byte
	keyReader := hkdf.New(sha256.New, append(dhShared[:]), nil, DH_HKDF_EPHEM)
	_, err := io.ReadFull(keyReader, derived[:])
	if err != nil {
		panic(err)
	}

	// Decapsulate sub shared-secrets
	var nonce [24]byte
	copy(nonce[:], ciphertext)
	outDH, ok := secretbox.Open(nil, ciphertext[24:32+24+secretbox.Overhead], &nonce, &derived)
	if !ok {
		panic("failed to verify mac of encapsulated key")
	}

	var outKyber []byte
	local.PrivkeyPQ.DecryptTo(outKyber, ciphertext[32+24+secretbox.Overhead:])

	// Derive shared secret
	h := hmac.New(sha256.New, append(outDH[:], outKyber[:]...))
	h.Write(EPHEM_HMAC)
	var shared MessageKey
	copy(shared[:], h.Sum(nil))

	return shared
}

// GenerateFingerprint takes a shared secret and the public values and turns
// it into a string of 9 groups of 4 base-10 digits.
func GenerateFingerprint(local *EphemPriv, remote *EphemPub, secret []byte) string {
	var localPub x25519.Key
	x25519.KeyGen(&localPub, &local.Privkey)

	synthpub := &EphemPub{
		Pubkey:   localPub,
		PubkeyPQ: local.PubkeyPQ,
	}

	b := new(bytes.Buffer)
	synthpub.Marshal(b)
	remote.Marshal(b)
	b.Write(secret)

	hash := argon2.IDKey(b.Bytes(), EPHEM_FINGERPRINT_SALT, 1, 64*1024, 1, 15)

	var out string
	in := new(big.Int).SetBytes(hash)
	for i := 0; i < 36; i++ {
		digit := new(big.Int).Mod(in, big.NewInt(10)).Int64()
		in = new(big.Int).Div(in, big.NewInt(10))

		out += fmt.Sprint(digit)

		if i%4 == 3 {
			// Add spaces every 4 digits
			out += " "
		}
	}
	out = strings.TrimRight(out, " ")

	return out
}
