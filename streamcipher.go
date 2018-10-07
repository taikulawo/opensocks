package main

import (
	"crypto/rc4"
	"crypto/sha256"
)

func NewRC4(raw []byte)*rc4.Cipher{
	sh256 := sha256.New()
	_, err := sh256.Write(raw)
	if err != nil{
		panic(err)
	}

	sh256Res := sh256.Sum(nil)
	cipher, err := rc4.NewCipher(sh256Res)
	if err != nil{
		panic(err)
	}
	return cipher
}

