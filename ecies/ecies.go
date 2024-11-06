// Copyright (c) 2013 Kyle Isom <kyle@tyrfingr.is>
// Copyright (c) 2012 The Go Authors. All rights reserved.
//
// Redistribution and use in source and binary forms, with or without
// modification, are permitted provided that the following conditions are
// met:
//
//    * Redistributions of source code must retain the above copyright
// notice, this list of conditions and the following disclaimer.
//    * Redistributions in binary form must reproduce the above
// copyright notice, this list of conditions and the following disclaimer
// in the documentation and/or other materials provided with the
// distribution.
//    * Neither the name of Google Inc. nor the names of its
// contributors may be used to endorse or promote products derived from
// this software without specific prior written permission.
//
// THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS
// "AS IS" AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT
// LIMITED TO, THE IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR
// A PARTICULAR PURPOSE ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT
// OWNER OR CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL,
// SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT
// LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE,
// DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY
// THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
// (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE
// OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.

package ecies

import (
	"crypto/cipher"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/hmac"
	"crypto/subtle"
	"encoding/binary"
	"encoding/hex"
	"hash"
	"io"
	"math/big"
	// "user/pkg/util"
)

// PublicKey is a representation of an elliptic curve public key.
type PublicKey struct {
	X *big.Int
	Y *big.Int
	elliptic.Curve
	Params *Params
}

// ExportECDSA Export an ECIES public key as an ECDSA public key.
func (pub *PublicKey) ExportECDSA() *ecdsa.PublicKey {
	return &ecdsa.PublicKey{Curve: pub.Curve, X: pub.X, Y: pub.Y}
}

// ImportECDSAPublic Import an ECDSA public key as an ECIES public key.
func ImportECDSAPublic(pub *ecdsa.PublicKey) *PublicKey {
	return &PublicKey{
		X:      pub.X,
		Y:      pub.Y,
		Curve:  pub.Curve,
		Params: ParamsFromCurve(pub.Curve),
	}
}

// PrivateKey is a representation of an elliptic curve private key.
type PrivateKey struct {
	PublicKey
	D *big.Int
}

// ExportECDSA Export an ECIES private key as an ECDSA private key.
func (pk *PrivateKey) ExportECDSA() *ecdsa.PrivateKey {
	pub := &pk.PublicKey
	pubECDSA := pub.ExportECDSA()
	return &ecdsa.PrivateKey{PublicKey: *pubECDSA, D: pk.D}
}

// ImportECDSA Import an ECDSA private key as an ECIES private key.
func ImportECDSA(pk *ecdsa.PrivateKey) *PrivateKey {
	pub := ImportECDSAPublic(&pk.PublicKey)
	return &PrivateKey{*pub, pk.D}
}

// GenerateKey Generate an elliptic curve public / private keypair. If params is nil,
// the recommended default parameters for the key will be chosen.
func GenerateKey(rand io.Reader, curve elliptic.Curve, params *Params) (prv *PrivateKey, er Err) {
	pb, x, y, err := elliptic.GenerateKey(curve, rand)
	if err != nil {
		er = ErrEccGenErr
		return
	}
	prv = new(PrivateKey)
	prv.PublicKey.X = x
	prv.PublicKey.Y = y
	prv.PublicKey.Curve = curve
	prv.D = new(big.Int).SetBytes(pb)
	if params == nil {
		params = ParamsFromCurve(curve)
	}
	prv.PublicKey.Params = params
	return
}

// MaxSharedKeyLength returns the maximum length of the shared key the
// public key can produce.
func MaxSharedKeyLength(pub *PublicKey) int {
	return (pub.Curve.Params().BitSize + 7) / 8
}

// GenerateShared ECDH key agreement method used to establish secret keys for encryption.
func (pk *PrivateKey) GenerateShared(pub *PublicKey, skLen, macLen int) (sk []byte, err Err) {
	if pk.PublicKey.Curve != pub.Curve {
		return nil, ErrEccInvalidCurve
	}
	if skLen+macLen > MaxSharedKeyLength(pub) {
		return nil, ErrEccSharedKeyTooBig
	}

	x, _ := pub.Curve.ScalarMult(pub.X, pub.Y, pk.D.Bytes())
	if x == nil {
		return nil, ErrEccSharedKeyIsPointAtInfinity
	}

	sk = make([]byte, skLen+macLen)
	skBytes := x.Bytes()
	copy(sk[len(sk)-len(skBytes):], skBytes)
	return sk, nil
}

// NIST SP 800-56 Concatenation Key Derivation Function (see section 5.8.1).
func concatKDF(hash hash.Hash, z, s1 []byte, kdLen int) []byte {
	counterBytes := make([]byte, 4)
	k := make([]byte, 0, roundup(kdLen, hash.Size()))
	for counter := uint32(1); len(k) < kdLen; counter++ {
		binary.BigEndian.PutUint32(counterBytes, counter)
		hash.Reset()
		hash.Write(counterBytes)
		hash.Write(z)
		hash.Write(s1)
		k = hash.Sum(k)
	}
	return k[:kdLen]
}

// roundup rounds size up to the next multiple of blockSize.
func roundup(size, blockSize int) int {
	return size + blockSize - (size % blockSize)
}

// deriveKeys creates the encryption and MAC keys using concatKDF.
func deriveKeys(hash hash.Hash, z, s1 []byte, keyLen int) (Ke, Km []byte) {
	K := concatKDF(hash, z, s1, 2*keyLen)
	Ke = K[:keyLen]
	Km = K[keyLen:]
	hash.Reset()
	hash.Write(Km)
	Km = hash.Sum(Km[:0])
	return Ke, Km
}

// messageTag computes the MAC of a message (called the tag) as per
// SEC 1, 3.5.
func messageTag(hash func() hash.Hash, km, msg, shared []byte) []byte {
	mac := hmac.New(hash, km)
	mac.Write(msg)
	mac.Write(shared)
	tag := mac.Sum(nil)
	return tag
}

// Generate an initialisation vector for CTR mode.
func generateIV(params *Params, rand io.Reader) (iv []byte, er Err) {
	iv = make([]byte, params.BlockSize)
	_, err := io.ReadFull(rand, iv)
	if err != nil {
		er = ErrEccIVGen
	}
	return
}

// symEncrypt carries out CTR encryption using the block cipher specified in the
func symEncrypt(rand io.Reader, params *Params, key, m []byte) (ct []byte, er Err) {
	c, err := params.Cipher(key)
	if err != nil {
		er = ErrEccKeySize
		return
	}

	iv, er := generateIV(params, rand)
	if er != nil {
		return
	}
	ctr := cipher.NewCTR(c, iv)

	ct = make([]byte, len(m)+params.BlockSize)
	copy(ct, iv)
	ctr.XORKeyStream(ct[params.BlockSize:], m)
	return
}

// symDecrypt carries out CTR decryption using the block cipher specified in
// the parameters
func symDecrypt(params *Params, key, ct []byte) (m []byte, er Err) {
	c, err := params.Cipher(key)
	if err != nil {
		er = ErrEccKeySize
		return
	}

	ctr := cipher.NewCTR(c, ct[:params.BlockSize])

	m = make([]byte, len(ct)-params.BlockSize)
	ctr.XORKeyStream(m, ct[params.BlockSize:])
	return
}

// Encrypt encrypts a message using ECIES as specified in SEC 1, 5.1.
//
// s1 and s2 contain shared information that is not part of the resulting
// ciphertext. s1 is fed into key derivation, s2 is fed into the MAC. If the
// shared information parameters aren't being used, they should be nil.
func Encrypt(rand io.Reader, pub *PublicKey, m, s1, s2 []byte) (ct []byte, er Err) {
	params, err := pubKeyParams(pub)
	if err != nil {
		return nil, err
	}

	R, er := GenerateKey(rand, pub.Curve, params)
	if er != nil {
		return nil, er
	}

	z, err := R.GenerateShared(pub, params.KeyLen, params.KeyLen)
	if err != nil {
		return nil, err
	}

	hashed := params.Hash()
	Ke, Km := deriveKeys(hashed, z, s1, params.KeyLen)

	em, err := symEncrypt(rand, params, Ke, m)
	if err != nil || len(em) <= params.BlockSize {
		return nil, err
	}

	d := messageTag(params.Hash, Km, em, s2)

	Rb := elliptic.Marshal(pub.Curve, R.PublicKey.X, R.PublicKey.Y)
	ct = make([]byte, len(Rb)+len(em)+len(d))
	copy(ct, Rb)
	copy(ct[len(Rb):], em)
	copy(ct[len(Rb)+len(em):], d)
	return ct, nil
}

// Decrypt decrypts an ECIES ciphertext.
func (pk *PrivateKey) Decrypt(c, s1, s2 []byte) (m []byte, err Err) {
	if len(c) == 0 {
		return nil, ErrEccInvalidMessage
	}
	params, err := pubKeyParams(&pk.PublicKey)
	if err != nil {
		return nil, err
	}

	hashed := params.Hash()

	var (
		rLen   int
		hLen   = hashed.Size()
		mStart int
		mEnd   int
	)

	switch c[0] {
	case 2, 3, 4:
		rLen = (pk.PublicKey.Curve.Params().BitSize + 7) / 4
		if len(c) < (rLen + hLen + 1) {
			return nil, ErrEccInvalidMessage
		}
	default:
		return nil, ErrEccInvalidPublicKey
	}

	mStart = rLen
	mEnd = len(c) - hLen

	R := new(PublicKey)
	R.Curve = pk.PublicKey.Curve
	R.X, R.Y = elliptic.Unmarshal(R.Curve, c[:rLen])
	if R.X == nil {
		return nil, ErrEccInvalidPublicKey
	}

	z, err := pk.GenerateShared(R, params.KeyLen, params.KeyLen)
	if err != nil {
		return nil, err
	}
	Ke, Km := deriveKeys(hashed, z, s1, params.KeyLen)

	d := messageTag(params.Hash, Km, c[mStart:mEnd], s2)
	if subtle.ConstantTimeCompare(c[mEnd:], d) != 1 {
		return nil, ErrEccInvalidMessage
	}

	return symDecrypt(params, Ke, c[mStart:mEnd])
}

func PrivateFromString(hexKey string) (*PrivateKey, Err) {
	pk, err := hex.DecodeString(hexKey)
	if err != nil {
		return nil, ErrPrvKeyDecode
	}

	x, y := elliptic.P256().ScalarBaseMult(pk)

	return &PrivateKey{
		PublicKey: PublicKey{
			X:      x,
			Y:      y,
			Curve:  elliptic.P256(),
			Params: ParamsFromCurve(elliptic.P256()),
		},
		D: big.NewInt(0).SetBytes(pk),
	}, nil
}

func PublicFromString(hexKey string) (*PublicKey, Err) {
	pb, err := hex.DecodeString(hexKey)
	if err != nil {
		return nil, ErrPubKeyDecode
	}

	x, y := elliptic.UnmarshalCompressed(elliptic.P256(), pb)
	return &PublicKey{
		X:      x,
		Y:      y,
		Curve:  elliptic.P256(),
		Params: ParamsFromCurve(elliptic.P256()),
	}, nil
}

func (pk *PrivateKey) String() string {
	return hex.EncodeToString(pk.D.Bytes())
}

func (pub *PublicKey) String() string {
	return hex.EncodeToString(elliptic.MarshalCompressed(pub.Curve, pub.X, pub.Y))
}
