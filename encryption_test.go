package main

import (
  "testing"
  "crypto/rand"

  "github.com/stretchr/testify/assert"
)


func TestNewEncryption(t *testing.T) {
  assert := assert.New(t)
  encryption, err := NewEncryption([]byte("1234567890abcdefghijklmnopqrstuv"));
  assert.NotNil(encryption)
  assert.NoError(err)
}

func TestNewEncryptionErrorTooShortKey(t *testing.T) {
  assert := assert.New(t)
  encryption, err := NewEncryption([]byte{});
  assert.Nil(encryption)
  assert.Error(err)
}

func TestNewEncryptionErroLittleShortKey(t *testing.T) {
  assert := assert.New(t)
  encryption, err := NewEncryption([]byte("1234567890abcdefghijklmnopqrstu"))
  assert.Nil(encryption)
  assert.Error(err)
}

func TestNewEncryptionErroLittleLongKey(t *testing.T) {
  assert := assert.New(t)
  encryption, err := NewEncryption([]byte("1234567890abcdefghijklmnopqrstuvw"))
  assert.Nil(encryption)
  assert.Error(err)
}

func TestNewEncryptionErrorTooLongKey(t *testing.T) {
  assert := assert.New(t)
  encryption, err := NewEncryption([]byte("1234567890abcdefghijklmnopqrstuvwxyz"))
  assert.Nil(encryption)
  assert.Error(err)
}

func TestEncryptAndDecrypt(t *testing.T) {
  assert := assert.New(t)
  key := make([]byte, 32)
  plaintext := make([]byte, 768)
  _, _ = rand.Read(key)
  _, _ = rand.Read(plaintext)
  encryption, _ := NewEncryption(key)
  ciphertext, err := encryption.encrypt(plaintext)
  assert.NotNil(encryption)
  assert.NoError(err)
  decrypted, err := encryption.decrypt(ciphertext)
  assert.Equal(plaintext, decrypted)
}

func TestDecryptError(t *testing.T) {
  assert := assert.New(t)
  key := make([]byte, 32)
  ciphertext := make([]byte, 768)
  _, _ = rand.Read(key)
  _, _ = rand.Read(ciphertext)
  encryption, _ := NewEncryption(key)
  decrypted, err := encryption.decrypt(ciphertext)
  assert.Nil(decrypted)
  assert.Error(err)
}
