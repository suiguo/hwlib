package rsa

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"os"
)

// 创建RSA
func GenerateRSAKey(bits int) (pub []byte, pri []byte, err error) {
	// Generate an RSA keypair.
	rsaKey, err := rsa.GenerateKey(rand.Reader, bits)
	if err != nil {
		return nil, nil, err
	}
	priPKCS8Key, err := x509.MarshalPKCS8PrivateKey(rsaKey)
	if err != nil {
		return nil, nil, err
	}
	priPKCS8Block := &pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: priPKCS8Key,
	}
	pri = pem.EncodeToMemory(priPKCS8Block)
	pubPKCS8Key, err := x509.MarshalPKIXPublicKey(&rsaKey.PublicKey)
	if err != nil {
		return nil, nil, err
	}
	pubPKCS8Block := &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: pubPKCS8Key,
	}
	pub = pem.EncodeToMemory(pubPKCS8Block)
	return pub, pri, nil
}

func GenerateRSAKeyFile(bits int, file string) error {
	// Generate an RSA keypair.
	rsaKey, err := rsa.GenerateKey(rand.Reader, bits)
	if err != nil {
		return err
	}
	priPKCS8Key, err := x509.MarshalPKCS8PrivateKey(rsaKey)
	if err != nil {
		return err
	}
	priPKCS8Block := &pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: priPKCS8Key,
	}
	f, err := os.Create("./" + file)
	if err != nil {
		return err
	}
	defer f.Close()
	// os.OpenFile("./RsaPri", flag int, perm FileMode)
	err = pem.Encode(f, priPKCS8Block)
	if err != nil {
		return err
	}
	pubPKCS8Key, err := x509.MarshalPKIXPublicKey(&rsaKey.PublicKey)
	if err != nil {
		return err
	}
	pubPKCS8Block := &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: pubPKCS8Key,
	}
	f2, err2 := os.Create("./" + file + ".pub")
	if err2 != nil {
		return err
	}
	defer f2.Close()
	return pem.Encode(f2, pubPKCS8Block)
}

// SignRSA ras签名 返回签名信息后的base64 (sha256)
func SignRSA(hashtype crypto.Hash, data []byte, pri []byte) (string, error) {
	myHash := hashtype
	hashInstance := myHash.New()
	hashInstance.Write(data)
	hashed := hashInstance.Sum(nil)
	block, _ := pem.Decode(pri)
	if block == nil {
		return "", fmt.Errorf("block is nil")
	}
	// x509.ParsePKCSPrivateKey()
	pkcs8key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return "", err
	}
	privateKey, ok := pkcs8key.(*rsa.PrivateKey)
	if !ok {
		return "", fmt.Errorf("privatekey is error")
	}
	signature, err := rsa.SignPKCS1v15(rand.Reader, privateKey, crypto.SHA256, hashed)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(signature), nil
}

// data 签名原始数据  base64Sig签名后base64的数据  pub公钥（sha256）
func VerifyRSA(hashtype crypto.Hash, data []byte, base64Sig string, pub []byte) bool {
	bytes, err := base64.StdEncoding.DecodeString(base64Sig)
	if err != nil {
		return false
	}
	myHash := hashtype
	hashInstance := myHash.New()
	hashInstance.Write(data)
	hashed := hashInstance.Sum(nil)
	block, _ := pem.Decode(pub)
	if block == nil {
		return false
	}
	publicKeyInterface, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return false
	}
	publicKey, ok := publicKeyInterface.(*rsa.PublicKey)
	if !ok {
		return false
	}
	err = rsa.VerifyPKCS1v15(publicKey, myHash, hashed, bytes)
	return err == nil
}

func Encrypt(pub []byte, data []byte) ([]byte, error) {
	block, _ := pem.Decode(pub)
	if block == nil {
		return nil, fmt.Errorf("block is nil")
	}
	publicKeyInterface, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	publicKey, ok := publicKeyInterface.(*rsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("pub error")
	}
	return rsa.EncryptPKCS1v15(rand.Reader, publicKey, data)
}

func Decrypt(pri []byte, data []byte) ([]byte, error) {
	block, _ := pem.Decode(pri)
	if block == nil {
		return nil, fmt.Errorf("block is nil")
	}
	// x509.ParsePKCSPrivateKey()
	pkcs8key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	privateKey, ok := pkcs8key.(*rsa.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("private key error")
	}
	return rsa.DecryptPKCS1v15(rand.Reader, privateKey, data)
}
