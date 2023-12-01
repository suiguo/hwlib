package gjwt

import (
	"crypto"
	"crypto/ed25519"
	"fmt"
	"sync"
	"time"

	gojwt "github.com/golang-jwt/jwt/v5"
	jsoniter "github.com/json-iterator/go"
)

type MethodSigning string

var json = jsoniter.ConfigCompatibleWithStandardLibrary

const (
	MethodEd25519   MethodSigning = "MethodEd25519"
	MethodECDSA256  MethodSigning = "MethodECDSA256"
	MethodECDSA384  MethodSigning = "MethodECDSA384"
	MethodECDSA512  MethodSigning = "MethodECDSA512"
	MethodRSA256    MethodSigning = "MethodRSA256"
	MethodRSA384    MethodSigning = "MethodRSA384"
	MethodRSA512    MethodSigning = "MethodRSA512"
	MethodRSAPSS256 MethodSigning = "MethodRSAPSS256"
	MethodRSAPSS384 MethodSigning = "MethodRSAPSS384"
	MethodRSAPSS512 MethodSigning = "MethodRSAPSS512"
	MethodHMAC256   MethodSigning = "MethodHMAC256"
	MethodHMAC384   MethodSigning = "MethodHMAC384"
	MethodHMAC512   MethodSigning = "MethodHMAC512"
)

var signFunc = map[MethodSigning]gojwt.SigningMethod{
	MethodEd25519:   gojwt.SigningMethodEdDSA,
	MethodECDSA256:  gojwt.SigningMethodES256,
	MethodECDSA384:  gojwt.SigningMethodES384,
	MethodECDSA512:  gojwt.SigningMethodES512,
	MethodRSA256:    gojwt.SigningMethodRS256,
	MethodRSA384:    gojwt.SigningMethodRS384,
	MethodRSA512:    gojwt.SigningMethodRS512,
	MethodRSAPSS256: gojwt.SigningMethodPS256,
	MethodRSAPSS384: gojwt.SigningMethodPS384,
	MethodRSAPSS512: gojwt.SigningMethodPS512,
	MethodHMAC256:   gojwt.SigningMethodHS256,
	MethodHMAC384:   gojwt.SigningMethodHS384,
	MethodHMAC512:   gojwt.SigningMethodHS512,
}

type JwtUitls interface {
	VerifyAndMarshal(token string, jwtdata interface{}) (bool, error)
	Verify(token string) (bool, any, error)
	SetSecretPub(method MethodSigning, key []byte)
	SetSecretPriv(method MethodSigning, key []byte)
	SignData(signtype MethodSigning, data interface{}, expires time.Duration) (string, error)
}

func NewJWT() JwtUitls {
	return &utils{}
}

type utils struct {
	JwtUitls
	secretMap sync.Map
	privMap   sync.Map
}
type CommonClaims struct {
	Data interface{} `json:"data"`
	*gojwt.RegisteredClaims
}

// data 签名数据 expires多长时间过期
func (u *utils) SignData(signtype MethodSigning, data interface{}, expires time.Duration) (signrs string, signerr error) {
	defer func() {
		err := recover()
		if err != nil {
			signrs = ""
			signerr = err.(error)
		}
	}()
	signdata := &CommonClaims{
		Data: data,
		RegisteredClaims: &gojwt.RegisteredClaims{
			ExpiresAt: gojwt.NewNumericDate(time.Now().Add(expires)),
		},
	}
	signMethod, ok := signFunc[signtype]
	if !ok {
		return "", fmt.Errorf("not support sign type")
	}
	token := gojwt.NewWithClaims(signMethod, signdata)
	priv, ok := u.privMap.Load(signtype)
	if !ok {
		return "", fmt.Errorf("not set priv key")
	}
	var key interface{}
	// var err error
	switch signtype {
	case MethodEd25519:
		if p, ok := priv.([]byte); ok {
			key = ed25519.PrivateKey(p)
		} else {
			key = priv
		}
	case MethodECDSA256, MethodECDSA384, MethodECDSA512:
		if p, ok := priv.([]byte); ok {
			tmp, err := Parse.ParseEcdsaPriv(string(p))
			if err != nil {
				return "", err
			}
			key = tmp
			u.privMap.Store(signtype, tmp)
		} else {
			key = priv
		}
	case MethodRSA256, MethodRSA384, MethodRSA512, MethodRSAPSS256, MethodRSAPSS384, MethodRSAPSS512:
		if p, ok := priv.([]byte); ok {
			tmp, err := Parse.ParseRsaPriv(string(p))
			if err != nil {
				return "", err
			}
			key = tmp
			u.privMap.Store(signtype, tmp)
		} else {
			key = priv
		}
	case MethodHMAC256, MethodHMAC384, MethodHMAC512:
		if p, ok := priv.([]byte); ok {
			key = p
		} else {
			return "", fmt.Errorf("not found sign type")
		}
	default:
		return "", fmt.Errorf("not found sign type")
	}
	if key == nil {
		return "", fmt.Errorf("not found sign type")
	}

	return token.SignedString(key)
}
func (u *utils) getMethodEd25519Key() (ed25519.PublicKey, error) {
	data, ok := u.secretMap.Load(MethodEd25519)
	if !ok {
		return nil, fmt.Errorf("not set MethodEd25519 SecretKey")
	}
	d, ok := data.(ed25519.PublicKey)
	if ok {
		return d, nil
	}
	tmp, ok := data.([]byte)
	if !ok {
		return nil, fmt.Errorf("not set Ed25519 SecretKey")
	}
	u.secretMap.Store(MethodEd25519, ed25519.PublicKey(tmp))
	return ed25519.PublicKey(tmp), nil
}

func (u *utils) VerifyAndMarshal(token string, jwtdata interface{}) (verifyrs bool, verifyerr error) {
	defer func() {
		err := recover()
		if err != nil {
			verifyrs = false
			verifyerr = err.(error)
		}
	}()
	t, err := gojwt.Parse(token, func(t *gojwt.Token) (interface{}, error) {
		switch method := t.Method.(type) {
		case *gojwt.SigningMethodRSA:
			return u.getRSAKey(method)
		case *gojwt.SigningMethodEd25519:
			return u.getMethodEd25519Key()
		case *gojwt.SigningMethodECDSA:
			return u.getECDSAKey(method)
		case *gojwt.SigningMethodHMAC:
			return u.getHMAC(method)
		case *gojwt.SigningMethodRSAPSS:
			return u.getRSAPASSKey(method)
		default:
			return nil, fmt.Errorf("not support")
		}
	})
	if err != nil {
		return false, err
	}
	if claims, ok := t.Claims.(gojwt.MapClaims); ok && t.Valid {
		if jwtdata == nil {
			return true, nil
		}
		data, ok := claims["data"]
		if !ok {
			return true, nil
		}
		d, e := json.Marshal(data)
		if e != nil {
			return true, e
		}
		err := json.Unmarshal(d, jwtdata)
		if err != nil {
			return true, err
		}
		return true, nil
	}
	return false, fmt.Errorf("verify fail")
}

// 验证jwttoken 并返回原始数据
func (u *utils) Verify(token string) (verifyrs bool, data any, verifyerr error) {
	defer func() {
		err := recover()
		if err != nil {
			verifyrs = false
			verifyerr = err.(error)
		}
	}()
	t, err := gojwt.Parse(token, func(t *gojwt.Token) (interface{}, error) {
		switch method := t.Method.(type) {
		case *gojwt.SigningMethodRSA:
			return u.getRSAKey(method)
		case *gojwt.SigningMethodEd25519:
			return u.getMethodEd25519Key()
		case *gojwt.SigningMethodECDSA:
			return u.getECDSAKey(method)
		case *gojwt.SigningMethodHMAC:
			return u.getHMAC(method)
		case *gojwt.SigningMethodRSAPSS:
			return u.getRSAPASSKey(method)
		default:
			return nil, fmt.Errorf("not support")
		}
	})
	if err != nil {
		return false, nil, err
	}
	if claims, ok := t.Claims.(gojwt.MapClaims); ok && t.Valid {
		data, ok := claims["data"]
		if !ok {
			return true, nil, nil
		}
		return true, data, nil
	}
	return false, nil, fmt.Errorf("verify fail")
}
func (u *utils) SetSecretPub(method MethodSigning, key []byte) {
	u.secretMap.Store(method, key)
}

func (u *utils) SetSecretPriv(method MethodSigning, key []byte) {
	u.privMap.Store(method, key)
}
func (u *utils) getRSAKey(method *gojwt.SigningMethodRSA) (interface{}, error) {
	var data interface{}
	var ok bool
	var methodname MethodSigning
	switch method.Hash {
	case crypto.SHA256:
		data, ok = u.secretMap.Load(MethodRSA256)
		methodname = MethodRSA256
	case crypto.SHA384:
		data, ok = u.secretMap.Load(MethodRSA384)
		methodname = MethodRSA384
	case crypto.SHA512:
		data, ok = u.secretMap.Load(MethodRSA512)
		methodname = MethodRSA512
	default:
		return nil, fmt.Errorf("not support hash type")
	}
	if data == nil || !ok {
		return nil, fmt.Errorf("not set MethodRSA SecretKey")
	}
	if byteskey, ok := data.([]byte); ok {
		key, err := Parse.ParseRsaPub(string(byteskey))
		if err != nil {
			return nil, err
		}
		u.secretMap.Store(methodname, key)
		return key, nil
	}
	return data, nil
}
func (u *utils) getECDSAKey(method *gojwt.SigningMethodECDSA) (interface{}, error) {
	var pub interface{}
	var ok bool
	var methodname MethodSigning
	switch method.Hash {
	case crypto.SHA256:
		pub, ok = u.secretMap.Load(MethodECDSA256)
		methodname = MethodECDSA256
	case crypto.SHA384:
		pub, ok = u.secretMap.Load(MethodECDSA384)
		methodname = MethodECDSA384
	case crypto.SHA512:
		pub, ok = u.secretMap.Load(MethodECDSA512)
		methodname = MethodECDSA512
	default:
		return nil, fmt.Errorf("not set secret key")
	}
	if !ok {
		return nil, fmt.Errorf("not set secret key")
	}
	if str, ok := pub.([]byte); ok {
		pubkey, err := Parse.ParseEcdsaPub(string(str))
		if err != nil {
			return nil, err
		}
		u.secretMap.Store(methodname, pubkey)
		return pubkey, nil
	}
	return pub, nil
}
func (u *utils) getRSAPASSKey(method *gojwt.SigningMethodRSAPSS) (interface{}, error) {
	var data interface{}
	var ok bool
	var methodname MethodSigning
	switch method.Hash {
	case crypto.SHA256:
		data, ok = u.secretMap.Load(MethodRSAPSS256)
		methodname = MethodRSAPSS256
	case crypto.SHA384:
		data, ok = u.secretMap.Load(MethodRSAPSS384)
		methodname = MethodRSAPSS384
	case crypto.SHA512:
		data, ok = u.secretMap.Load(MethodRSAPSS512)
		methodname = MethodRSAPSS512
	default:
		return nil, fmt.Errorf("not support hash type")
	}
	if data == nil || !ok {
		return nil, fmt.Errorf("not set MethodRSA SecretKey")
	}
	if byteskey, ok := data.([]byte); ok {
		key, err := Parse.ParseRsaPub(string(byteskey))
		if err != nil {
			return nil, err
		}
		u.secretMap.Store(methodname, key)
		return key, nil
	}
	return data, nil
}
func (u *utils) getHMAC(method *gojwt.SigningMethodHMAC) (interface{}, error) {
	var data interface{}
	var ok bool
	switch method.Hash {
	case crypto.SHA256:
		data, ok = u.secretMap.Load(MethodHMAC256)
	case crypto.SHA384:
		data, ok = u.secretMap.Load(MethodHMAC384)
	case crypto.SHA512:
		data, ok = u.secretMap.Load(MethodHMAC512)
	default:
		return nil, fmt.Errorf("not support hash type")
	}
	if !ok || data == nil {
		return nil, fmt.Errorf("not set MethodRSA SecretKey")
	}
	if k, ok := data.([]byte); ok {
		return k, nil
	}
	return nil, fmt.Errorf("not set MethodRSA SecretKey")
}

// func SignToken(data []byte,gojwt.SigningMethodECDSA,)
