package crypt

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"fmt"
)

func GenPrivateKey() *rsa.PrivateKey {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		fmt.Println(err.Error)
	}
	return privateKey
}

func GetPublicKey(privateKey *rsa.PrivateKey) *rsa.PublicKey {
	return &privateKey.PublicKey
}

func Encrypt(publicKey *rsa.PublicKey, message []byte) []byte {
	sha256 := sha256.New()
	label := []byte("")
	encryptedmsg, err := rsa.EncryptOAEP(sha256, rand.Reader, publicKey, message, label)
	if err != nil {
		fmt.Println(err.Error)
	}
	return encryptedmsg
}

func Decrypt(privateKey *rsa.PrivateKey, message []byte) []byte {
	label := []byte("")
	sha256 := sha256.New()
	decryptedmsg, err := rsa.DecryptOAEP(sha256, nil, privateKey, message, label)
	if err != nil {
		fmt.Println(err.Error)
	}
	return decryptedmsg
}
