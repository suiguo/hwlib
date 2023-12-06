package gjwt

import (
	"log"
	"testing"
	"time"

	gojwt "github.com/golang-jwt/jwt/v5"
)

type Data struct {
	Food string `json:"foo"`
	Nbf  int64  `json:"nbf"`
	gojwt.RegisteredClaims
}

func Ecdsa256() bool {
	pubstr, pristr, err := Gen.GenEcdsa(CurveTypeP256)
	if err != nil {
		panic(err)
	}
	tool := NewJWT()
	tool.SetSecretPub(MethodECDSA256, []byte(pubstr))
	tool.SetSecretPriv(MethodECDSA256, []byte(pristr))
	d2 := &CommonClaims{}
	out2, err := tool.SignData(MethodECDSA256, "1234", time.Second*3)
	if err != nil {
		panic(err)
	}
	tool.VerifyAndMarshal(out2, d2)
	tool.VerifyAndMarshal(out2, d2)
	tool.VerifyAndMarshal(out2, d2)
	ok2, err2 := tool.VerifyAndMarshal(out2, d2)
	return ok2 && err2 == nil && d2.Data == "1234"
}

func Ecdsa512() bool {
	pubstr, pristr, err := Gen.GenEcdsa(CurveTypeP521)
	if err != nil {
		panic(err)
	}
	tool := NewJWT()
	tool.SetSecretPub(MethodECDSA512, []byte(pubstr))
	tool.SetSecretPriv(MethodECDSA512, []byte(pristr))
	d2 := &CommonClaims{}
	out2, err := tool.SignData(MethodECDSA512, "1234", time.Second*3)
	if err != nil {
		panic(err)
	}
	tool.VerifyAndMarshal(out2, d2)
	tool.VerifyAndMarshal(out2, d2)
	tool.VerifyAndMarshal(out2, d2)
	ok2, err2 := tool.VerifyAndMarshal(out2, d2)
	return ok2 && err2 == nil && d2.Data == "1234"
}

func Ecdsa384() bool {
	pubstr, pristr, err := Gen.GenEcdsa(CurveTypeP384)
	if err != nil {
		panic(err)
	}
	tool := NewJWT()
	tool.SetSecretPub(MethodECDSA384, []byte(pubstr))
	tool.SetSecretPriv(MethodECDSA384, []byte(pristr))
	d2 := &CommonClaims{}
	out2, err := tool.SignData(MethodECDSA384, "1234", time.Second*3)
	if err != nil {
		panic(err)
	}
	tool.VerifyAndMarshal(out2, d2)
	tool.VerifyAndMarshal(out2, d2)
	tool.VerifyAndMarshal(out2, d2)
	ok2, err2 := tool.VerifyAndMarshal(out2, d2)
	return ok2 && err2 == nil && d2.Data == "1234"
}
func Ed25519() bool {
	pub, pri, err := Gen.GenED25519()
	if err != nil {
		panic(err)
	}
	tool := NewJWT()
	tool.SetSecretPub(MethodEd25519, pub)
	tool.SetSecretPriv(MethodEd25519, []byte(pri))
	d2 := &CommonClaims{}
	out2, err := tool.SignData(MethodEd25519, "1234", time.Second*3)
	if err != nil {
		panic(err)
	}
	tool.VerifyAndMarshal(out2, d2)
	tool.VerifyAndMarshal(out2, d2)
	tool.VerifyAndMarshal(out2, d2)
	ok2, err2 := tool.VerifyAndMarshal(out2, d2)
	return ok2 && err2 == nil && d2.Data == "1234"
}

func Rsa256() bool {
	pub, pri, err := Gen.GenRsa(1024)
	if err != nil {
		panic(err)
	}
	tool := NewJWT()
	tool.SetSecretPub(MethodRSA256, []byte(pub))
	tool.SetSecretPriv(MethodRSA256, []byte(pri))
	d2 := &CommonClaims{}
	out2, err := tool.SignData(MethodRSA256, "1234", time.Second*3)
	if err != nil {
		panic(err)
	}
	tool.VerifyAndMarshal(out2, d2)
	tool.VerifyAndMarshal(out2, d2)
	tool.VerifyAndMarshal(out2, d2)
	ok2, err2 := tool.VerifyAndMarshal(out2, d2)
	return ok2 && err2 == nil && d2.Data == "1234"
}

/*
MethodRSA256   MethodSigning = "MethodRSA256"
MethodRSA384   MethodSigning = "MethodRSA384"
MethodRSA512   MethodSigning = "MethodRSA512"
*/
func Rsa512() bool {
	pub, pri, err := Gen.GenRsa(1024)
	if err != nil {
		panic(err)
	}
	tool := NewJWT()
	tool.SetSecretPub(MethodRSA512, []byte(pub))
	tool.SetSecretPriv(MethodRSA512, []byte(pri))
	d2 := &CommonClaims{}
	out2, err := tool.SignData(MethodRSA512, "1234", time.Second*3)
	if err != nil {
		panic(err)
	}
	tool.VerifyAndMarshal(out2, d2)
	tool.VerifyAndMarshal(out2, d2)
	tool.VerifyAndMarshal(out2, d2)
	ok2, err2 := tool.VerifyAndMarshal(out2, d2)
	return ok2 && err2 == nil && d2.Data == "1234"
}

func Rsa384() bool {
	pub, pri, err := Gen.GenRsa(1024)
	if err != nil {
		panic(err)
	}
	tool := NewJWT()
	tool.SetSecretPub(MethodRSA384, []byte(pub))
	tool.SetSecretPriv(MethodRSA384, []byte(pri))
	d2 := &CommonClaims{}
	out2, err := tool.SignData(MethodRSA384, "1234", time.Second*3)
	if err != nil {
		panic(err)
	}
	tool.VerifyAndMarshal(out2, d2)
	tool.VerifyAndMarshal(out2, d2)
	tool.VerifyAndMarshal(out2, d2)
	ok2, err2 := tool.VerifyAndMarshal(out2, d2)
	return ok2 && err2 == nil && d2.Data == "1234"
}

func RsaPss256() bool {
	pub, pri, err := Gen.GenRsa(1024)
	if err != nil {
		panic(err)
	}
	tool := NewJWT()
	tool.SetSecretPub(MethodRSAPSS256, []byte(pub))
	tool.SetSecretPriv(MethodRSAPSS256, []byte(pri))
	d2 := &CommonClaims{}
	out2, err := tool.SignData(MethodRSAPSS256, "1234", time.Second*3)
	if err != nil {
		panic(err)
	}
	tool.VerifyAndMarshal(out2, d2)
	tool.VerifyAndMarshal(out2, d2)
	tool.VerifyAndMarshal(out2, d2)
	ok2, err2 := tool.VerifyAndMarshal(out2, d2)
	return ok2 && err2 == nil && d2.Data == "1234"
}

func RsaPss512() bool {
	pub, pri, err := Gen.GenRsa(2048)
	if err != nil {
		panic(err)
	}
	tool := NewJWT()
	tool.SetSecretPub(MethodRSAPSS512, []byte(pub))
	tool.SetSecretPriv(MethodRSAPSS512, []byte(pri))
	d2 := &CommonClaims{}
	out2, err := tool.SignData(MethodRSAPSS512, "1234", time.Second*3)
	if err != nil {
		panic(err)
	}
	tool.VerifyAndMarshal(out2, d2)
	tool.VerifyAndMarshal(out2, d2)
	tool.VerifyAndMarshal(out2, d2)
	ok2, err2 := tool.VerifyAndMarshal(out2, d2)
	return ok2 && err2 == nil && d2.Data == "1234"
}

func RsaPss384() bool {
	pub, pri, err := Gen.GenRsa(1024)
	if err != nil {
		panic(err)
	}
	tool := NewJWT()
	tool.SetSecretPub(MethodRSAPSS384, []byte(pub))
	tool.SetSecretPriv(MethodRSAPSS384, []byte(pri))
	d2 := &CommonClaims{}
	out2, err := tool.SignData(MethodRSAPSS384, "1234", time.Second*3)
	if err != nil {
		panic(err)
	}
	tool.VerifyAndMarshal(out2, d2)
	tool.VerifyAndMarshal(out2, d2)
	tool.VerifyAndMarshal(out2, d2)
	ok2, err2 := tool.VerifyAndMarshal(out2, d2)
	return ok2 && err2 == nil && d2.Data == "1234"
}
func TestTool(t *testing.T) {
	if Ecdsa256() && Ecdsa384() && Ecdsa512() {
		log.Println("ecdsa pass")
	} else {
		log.Fatal("ecdsa fail")
	}
	// return
	if Ed25519() {
		log.Println("ed25519 pass")
	} else {
		log.Fatal("ed25519 fail")
	}
	if Rsa256() {
		log.Println("rsa Rsa256 pass")
	} else {
		log.Fatal("rsa Rsa256 fail")
	}
	if Rsa384() {
		log.Println("rsa Rsa384 pass")
	} else {
		log.Fatal("rsa Rsa384 fail")
	}
	if Rsa512() {
		log.Println("rsa Rsa512 pass")
	} else {
		log.Fatal("rsa Rsa512 fail")
	}
	if RsaPss256() {
		log.Println("RsaPss256 pass")
	} else {
		log.Fatal("RsaPss256 fail")
	}
	if RsaPss384() {
		log.Println("RsaPss384 pass")
	} else {
		log.Fatal("RsaPss384 fail")
	}
	if RsaPss512() {
		log.Println("RsaPss512 pass")
	} else {
		log.Fatal("RsaPss512 fail")
	}
	// return
}
