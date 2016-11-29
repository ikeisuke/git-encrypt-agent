package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"log"
)

type Encryption struct {
	block cipher.Block
}

func NewEncryption(key []byte) (*Encryption, error) {
	e := new(Encryption)
	block, err := aes.NewCipher(key[:32])
	if err != nil {
		log.Printf("19, %v", err)
		return nil, err
	}
	e.block = block
	return e, nil
}

func (e *Encryption) encrypt(value []byte) ([]byte, error) {
	iv := make([]byte, e.block.BlockSize())
	if _, err := rand.Read(iv); err != nil {
		return nil, err
	}
	stream := cipher.NewCTR(e.block, iv)
	encrypted := make([]byte, len(value))
	stream.XORKeyStream(encrypted, value)
	ciphertext := make([]byte, 0, len(encrypted)+len(iv))
	ciphertext = append(ciphertext, iv...)
	ciphertext = append(ciphertext, encrypted...)
	encoded := base64.StdEncoding.EncodeToString(ciphertext)
	return []byte(encoded), nil
}

func (e *Encryption) decrypt(value []byte) ([]byte, error) {
	decoded, err := base64.StdEncoding.DecodeString(string(value))
	if err != nil {
		return nil, err
	}
	encrypted := []byte(decoded)
	iv := encrypted[:e.block.BlockSize()]
	ciphertext := encrypted[e.block.BlockSize():]
	stream := cipher.NewCTR(e.block, iv)
	plain := make([]byte, len(ciphertext))
	stream.XORKeyStream(plain, ciphertext)
	return plain, nil
}
