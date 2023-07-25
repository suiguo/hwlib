package gjwt

import (
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/rand"
	grsa "crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"

	"github.com/suiguo/hwlib/rsa"
)

// 生成密钥函数
type GenEcdsaFunc func(ctype CurveType) (ecdspub_base64 string, ecdspri_base64 string, ecderr error)
type GenED25519Func func() (pub ed25519.PublicKey, pri ed25519.PrivateKey, err error)
type GenRsaFunc func(bits int) (pub []byte, pri []byte, err error)
type GenHmacFunc func(bits int) (key []byte, err error)

// 解析密钥函数
type ParseEcdsaPubFunc func(base64pub string) (*ecdsa.PublicKey, error)
type ParseEcdsaPrivFunc func(base64priv string) (*ecdsa.PrivateKey, error)
type ParseRsaPubFunc func(pub string) (*grsa.PublicKey, error)
type ParseRsaPrivFunc func(priv string) (*grsa.PrivateKey, error)

func parseRsaPub(pub string) (*grsa.PublicKey, error) {
	block, _ := pem.Decode([]byte(pub))
	if block == nil {
		return nil, fmt.Errorf("block is nil")
	}
	pkcs8key, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	k, ok := pkcs8key.(*grsa.PublicKey)
	if ok {
		return k, nil
	}
	return nil, fmt.Errorf("parse error")
}

func parseRsaPriv(priv string) (*grsa.PrivateKey, error) {
	block, _ := pem.Decode([]byte(priv))
	if block == nil {
		return nil, fmt.Errorf("block is nil")
	}
	pkcs8key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	k, ok := pkcs8key.(*grsa.PrivateKey)
	if ok {
		return k, nil
	}
	return nil, fmt.Errorf("parse error")
}

func hmac(bits int) ([]byte, error) {
	key := make([]byte, bits)
	n, err := rand.Read(key)
	if err != nil {
		return nil, err
	}
	if n != bits {
		return nil, fmt.Errorf("gen hmac error")
	}
	return key, err
}

var Gen = struct {
	GenEcdsa   GenEcdsaFunc
	GenED25519 GenED25519Func
	GenRsa     GenRsaFunc
	GenHmac    GenHmacFunc
}{
	GenEcdsa:   generateEcdsa,
	GenED25519: generateEd25519,
	GenRsa:     rsa.GenerateRSAKey,
	GenHmac:    hmac,
}

var Parse = struct {
	ParseEcdsaPub  ParseEcdsaPubFunc
	ParseEcdsaPriv ParseEcdsaPrivFunc
	ParseRsaPub    ParseRsaPubFunc
	ParseRsaPriv   ParseRsaPrivFunc
}{
	ParseEcdsaPub:  parseEcdsaPub,
	ParseEcdsaPriv: parseEcdsaPri,
	ParseRsaPub:    parseRsaPub,
	ParseRsaPriv:   parseRsaPriv,
}
