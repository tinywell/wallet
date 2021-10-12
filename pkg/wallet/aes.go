package wallet

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"errors"
	"io"
)

const (
	// AESSize 密钥长度
	AESSize = 16
)

// AESEncrypt 加密
func AESEncrypt(plaintext []byte, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(preKey(key))
	if err != nil {
		return nil, err
	}

	ciphertext := make([]byte, aes.BlockSize+len(plaintext))
	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return nil, err
	}
	mode := cipher.NewCFBEncrypter(block, iv)
	mode.XORKeyStream(ciphertext[aes.BlockSize:], plaintext)
	return ciphertext, nil
}

// AESDecrypt 解密
func AESDecrypt(ciphertext []byte, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(preKey(key))
	if err != nil {
		return nil, err
	}
	if len(ciphertext) < aes.BlockSize {
		return nil, errors.New("ciphertext too short")
	}
	iv := ciphertext[:aes.BlockSize]
	ciphertext = ciphertext[aes.BlockSize:]

	mode := cipher.NewCFBDecrypter(block, iv)
	mode.XORKeyStream(ciphertext, ciphertext)
	return ciphertext, nil
}

func preKey(key []byte) []byte {
	hash := sha256.New()
	hash.Write(key)
	hk := hash.Sum(nil)
	return hk[:AESSize]
}
