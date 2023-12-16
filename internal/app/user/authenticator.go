package user

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
)

type Authenticator struct {
	SecretKeyStore interface{ Secret() []byte }
}

var ErrAuthenticationFailed = errors.New("authenticaion failed")

func (s Authenticator) Authenticate(token string) (*User, error) {
	data, err := hex.DecodeString(token)
	if err != nil {
		return nil, fmt.Errorf("auth: %w", err)
	}

	secret := sha256.Sum256(s.SecretKeyStore.Secret())
	key := secret[:]

	if len(data) < signatureSize {
		return nil, ErrAuthenticationFailed
	}

	encrSign := data[len(data)-signatureSize:]
	id := data[0 : len(data)-signatureSize]

	sign, err := decrypt(encrSign, key)
	if err != nil {
		return nil, ErrAuthenticationFailed
	}

	h := hmac.New(sha256.New, key)
	h.Write(id)
	hash := h.Sum(nil)

	if hmac.Equal(sign, hash) {
		return &User{UserID: string(id)}, nil
	}

	return nil, ErrAuthenticationFailed
}

func (s Authenticator) Sign(token string) (string, error) {
	data := []byte(token)

	secret := sha256.Sum256(s.SecretKeyStore.Secret())
	key := secret[:]

	h := hmac.New(sha256.New, key)
	h.Write(data)
	hash := h.Sum(nil)

	signature, err := encrypt(hash, key)
	if err != nil {
		return "", fmt.Errorf("auth: failed to encrypt hash: %w", err)
	}

	signed := append(data, signature...)

	return hex.EncodeToString(signed), nil
}

var signatureSize = sha256.Size + aes.BlockSize

func encrypt(data, key []byte) ([]byte, error) {
	aesgcm, nonce, err := initCipher(key)
	if err != nil {
		return nil, err
	}

	return aesgcm.Seal(nil, nonce, data, nil), nil
}

func decrypt(data, key []byte) ([]byte, error) {
	aesgcm, nonce, err := initCipher(key)
	if err != nil {
		return nil, err
	}

	return aesgcm.Open(nil, nonce, data, nil)
}

func initCipher(key []byte) (cipher.AEAD, []byte, error) {
	aesblock, err := aes.NewCipher(key)
	if err != nil {
		return nil, nil, err
	}

	aesgcm, err := cipher.NewGCM(aesblock)
	if err != nil {
		return nil, nil, err
	}

	nonce := key[len(key)-aesgcm.NonceSize():]

	return aesgcm, nonce, nil
}

type KeyStore struct {
	keyFunc func() []byte
}

func NewKeyStore(f func() []byte) KeyStore {
	return KeyStore{keyFunc: f}
}

func (s KeyStore) Secret() []byte {
	return s.keyFunc()
}
