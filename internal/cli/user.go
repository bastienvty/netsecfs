package cli

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"

	"github.com/bastienvty/netsecfs/internal/crypto"
	"github.com/bastienvty/netsecfs/internal/db/meta"
	"golang.org/x/crypto/argon2"
)

type params struct {
	memory      uint32
	iterations  uint32
	parallelism uint8
	saltLength  uint32
	keyLength   uint32
}

const (
	DefaultMemory      = 64 * 1024
	DefaultIterations  = 3
	DefaultParallelism = 2
	DefaultSaltLength  = 16
	DefaultKeyLength   = 32
)

func defaultParams() *params {
	return &params{
		memory:      DefaultMemory,
		iterations:  DefaultIterations,
		parallelism: DefaultParallelism,
		saltLength:  DefaultSaltLength,
		keyLength:   DefaultKeyLength,
	}
}

type User struct {
	username string
	password string

	m          meta.Meta
	enc        crypto.CryptoHelper
	PrivateKey *rsa.PrivateKey
	masterKey  []byte
	rootKey    []byte
}

func (u *User) checkUser() bool {
	err := u.m.CheckUser(u.username)
	return err != 0
}

func (u *User) CreateUser() bool {
	if u.username == "" || u.password == "" {
		return false
	}
	if u.checkUser() {
		return false
	}
	p := defaultParams()
	salt := make([]byte, p.saltLength)
	_, err := rand.Read(salt)
	if err != nil {
		return false
	}
	masterKey := argon2.IDKey([]byte(u.password), salt, p.iterations, p.memory, p.parallelism, p.keyLength)

	// b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	hashMaster := sha256.New()
	_, err = hashMaster.Write(masterKey)
	if err != nil {
		return false
	}
	hashMasterKey := hashMaster.Sum(nil)
	privKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return false
	}
	pubKey := &privKey.PublicKey
	privKeyBytes := x509.MarshalPKCS1PrivateKey(privKey)
	pubKeyBytes := x509.MarshalPKCS1PublicKey(pubKey)

	rootKey := make([]byte, p.keyLength)
	_, err = rand.Read(rootKey)
	if err != nil {
		return false
	}

	masterCipher, ok := u.enc.Encrypt(masterKey, rootKey)
	if ok != nil {
		return false
	}
	privCipher, ok := u.enc.Encrypt(masterKey, privKeyBytes)
	if ok != nil {
		return false
	}

	errno := u.m.CreateUser(u.username, hashMasterKey, salt, masterCipher, privCipher, pubKeyBytes)
	if errno != 0 {
		return false
	}

	u.masterKey = masterKey
	u.rootKey = rootKey
	u.PrivateKey = privKey
	return true
}

func (u *User) VerifyUser() bool {
	if u.username == "" || u.password == "" {
		return false
	}
	p := defaultParams()
	var salt []byte
	errno := u.m.GetSalt(u.username, &salt)
	if errno != 0 {
		return false
	}
	masterKey := argon2.IDKey([]byte(u.password), salt, p.iterations, p.memory, p.parallelism, p.keyLength)

	hashMaster := sha256.New()
	_, err := hashMaster.Write(masterKey)
	if err != nil {
		return false
	}
	hashMasterKey := hashMaster.Sum(nil)
	var masterCipher, privCipher []byte
	errno = u.m.VerifyUser(u.username, hashMasterKey, &masterCipher, &privCipher)
	if errno != 0 {
		return false
	}

	rootKey, ok := u.enc.Decrypt(masterKey, masterCipher)
	if ok != nil {
		return false
	}
	privKeyBytes, ok := u.enc.Decrypt(masterKey, privCipher)
	if ok != nil {
		return false
	}

	privKey, err := x509.ParsePKCS1PrivateKey(privKeyBytes)
	if err != nil {
		return false
	}

	u.masterKey = masterKey
	u.rootKey = rootKey
	u.PrivateKey = privKey
	return true
}
