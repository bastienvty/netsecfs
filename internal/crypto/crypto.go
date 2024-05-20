package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"io"
)

type Crypto interface {
	Encrypt(key, plaintext []byte) ([]byte, error)
	Decrypt(key, ciphertext []byte) ([]byte, error)
}

type CryptoHelper struct {
}

func (c *CryptoHelper) Encrypt(key, plaintext []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		panic(err.Error())
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	// Never use more than 2^32 random nonces with a given key because of the risk of a repeat.
	nonce := make([]byte, aesgcm.NonceSize()) // 12 bytes by default
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	// encrypt an prepend the nonce to the ciphertext before returning it
	ciphertext := aesgcm.Seal(nonce, nonce, plaintext, nil)

	return ciphertext, nil
}

func (c *CryptoHelper) Decrypt(key, ciphertext []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	// the nonce is prepended to the cipher text so we need to make sure it is still there and length matches up
	nonceSize := aesgcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, err
	}

	// now we split the nonce from the ciptertext
	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]

	plaintext, err := aesgcm.Open(nil, nonce, ciphertext, nil)

	return plaintext, err
}
