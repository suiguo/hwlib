package gjwt

import (
	"crypto/ed25519"
	"crypto/rand"
)

func generateEd25519() (pub ed25519.PublicKey, pri ed25519.PrivateKey, err error) {
	return ed25519.GenerateKey(rand.Reader)
}
