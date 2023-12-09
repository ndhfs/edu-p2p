package codec

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
)

var (
	key = "1234567891234567"
)

type Json[T any] struct {
}

func NewJson[T any]() *Json[T] {
	return &Json[T]{}
}

func (j *Json[T]) Encode(t T) ([]byte, error) {
	b, err := json.Marshal(t)
	if err != nil {
		return nil, fmt.Errorf("error marshal to json. %w", err)
	}
	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		return nil, err
	}
	ciphertext := make([]byte, aes.BlockSize+len(b))
	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return nil, err
	}
	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream(ciphertext[aes.BlockSize:], b)
	return ciphertext, nil
}

func (j *Json[T]) Decode(bytes []byte) (T, error) {
	// Инициализация блочного шифра
	block, _ := aes.NewCipher([]byte(key))
	// Расшифровка текста
	iv := bytes[:aes.BlockSize]
	ciphertextBytes := bytes[aes.BlockSize:]
	cfb := cipher.NewCFBDecrypter(block, iv)
	decrypted := make([]byte, len(ciphertextBytes))
	cfb.XORKeyStream(decrypted, ciphertextBytes)

	var v T
	err := json.Unmarshal(decrypted, &v)
	return v, err
}
