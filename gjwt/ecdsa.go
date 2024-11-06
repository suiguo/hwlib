package gjwt

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"math/big"
)

type CurveType string

const (
	CurveTypeP224 CurveType = "P224"
	CurveTypeP256 CurveType = "P256"
	CurveTypeP384 CurveType = "P384"
	CurveTypeP521 CurveType = "P521"
)

var curveMap = map[CurveType]elliptic.Curve{
	CurveTypeP224: elliptic.P224(),
	CurveTypeP256: elliptic.P256(),
	CurveTypeP384: elliptic.P384(),
	CurveTypeP521: elliptic.P521(),
}

func intToBytes(n int) [4]byte {
	x := int32(n)
	bytesBuffer := bytes.NewBuffer([]byte{})
	binary.Write(bytesBuffer, binary.BigEndian, x)
	bytes := bytesBuffer.Bytes()
	out := [4]byte{}
	for i := 0; i < 4; i++ {
		out[i] = bytes[i]
	}
	return out
}
func bytesToInt(b []byte) int {
	bytesBuffer := bytes.NewBuffer(b)
	var x int32
	binary.Read(bytesBuffer, binary.BigEndian, &x)
	return int(x)
}
func generateEcdsa(ctype CurveType) (ecdspub string, ecdspri string, ecderr error) {
	curve, ok := curveMap[ctype]
	if !ok {
		return "", "", fmt.Errorf("not support CurveType")
	}
	ecdsapri, err := ecdsa.GenerateKey(curve, rand.Reader)
	if err != nil {
		return "", "", err
	}
	pri, err := x509.MarshalECPrivateKey(ecdsapri)
	if err != nil {
		return "", "", err
	}
	x := ecdsapri.PublicKey.X.Bytes()
	y := ecdsapri.PublicKey.Y.Bytes()
	pubbytes := make([]byte, 0)
	ctypebytes := []byte(ctype)
	length := intToBytes(len(ctypebytes))
	pubbytes = append(pubbytes, length[0:4]...)
	pubbytes = append(pubbytes, ctypebytes...)
	///
	length = intToBytes(len(x))
	pubbytes = append(pubbytes, length[0:4]...)
	pubbytes = append(pubbytes, x...)
	//
	length = intToBytes(len(y))
	pubbytes = append(pubbytes, length[0:4]...)
	pubbytes = append(pubbytes, y...)
	return base64.StdEncoding.EncodeToString(pubbytes), base64.StdEncoding.EncodeToString(pri), nil
}

func parseEcdsaPri(base64pri string) (*ecdsa.PrivateKey, error) {
	pri, err := base64.StdEncoding.DecodeString(base64pri)
	if err != nil {
		return nil, err
	}
	ecpri, err := x509.ParseECPrivateKey(pri)
	if err != nil {
		return nil, err
	}
	return ecpri, err
}
func parseEcdsaPub(base64pub string) (*ecdsa.PublicKey, error) {
	pub, err := base64.StdEncoding.DecodeString(base64pub)
	if err != nil {
		return nil, err
	}
	//解析curvetype
	if len(pub) < 4 {
		return nil, fmt.Errorf("formate error")
	}
	idx := 0
	lenbytes := pub[idx : idx+4]
	length := bytesToInt(lenbytes)
	if length < 0 {
		return nil, fmt.Errorf("formate error")
	}
	idx = idx + 4
	if len(pub) < idx+length {
		return nil, fmt.Errorf("formate error")
	}
	ctype := string(pub[idx : idx+length])
	curve, ok := curveMap[CurveType(ctype)]
	if !ok {
		return nil, fmt.Errorf("curve not exist")
	}
	//解析 x
	idx = idx + length
	if len(pub) < idx+4 {
		return nil, fmt.Errorf("formate error")
	}
	lenbytes = pub[idx : idx+4]
	length = bytesToInt(lenbytes)
	if length < 0 {
		return nil, fmt.Errorf("formate error")
	}
	idx = idx + 4
	if len(pub) < idx+length {
		return nil, fmt.Errorf("formate error")
	}
	bytex := pub[idx : idx+length]
	x := new(big.Int).SetBytes(bytex)
	//解析 y
	idx = idx + length
	if len(pub) < idx+4 {
		return nil, fmt.Errorf("formate error")
	}
	lenbytes = pub[idx : idx+4]
	length = bytesToInt(lenbytes)
	if length < 0 {
		return nil, fmt.Errorf("formate error")
	}
	idx = idx + 4
	if len(pub) < idx+length {
		return nil, fmt.Errorf("formate error")
	}
	bytey := pub[idx : idx+length]
	y := new(big.Int).SetBytes(bytey)
	return &ecdsa.PublicKey{
		Curve: curve,
		X:     x,
		Y:     y,
	}, nil
}
