package cli

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"fmt"
	"os"
	"syscall"

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
	DefaultMemory      = 512 * 1024 // 512 MB
	DefaultIterations  = 5          // increase time to login
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
	enc        crypto.Crypto
	privateKey *rsa.PrivateKey
	masterKey  []byte
	rootKey    []byte
}

func (u *User) checkUser() bool {
	err := u.m.CheckUser(u.username)
	return err != 0
}

func (u *User) createUser() bool {
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

	rootCipher, ok := u.enc.Encrypt(masterKey, rootKey)
	if ok != nil {
		return false
	}
	privCipher, ok := u.enc.Encrypt(masterKey, privKeyBytes)
	if ok != nil {
		return false
	}

	errno := u.m.CreateUser(u.username, hashMasterKey, salt, rootCipher, privCipher, pubKeyBytes)
	if errno != 0 {
		return false
	}

	u.masterKey = masterKey
	u.rootKey = rootKey
	u.privateKey = privKey
	return true
}

func (u *User) verifyUser() bool {
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
	var rootCipher, privCipher []byte
	errno = u.m.VerifyUser(u.username, hashMasterKey, &rootCipher, &privCipher)
	if errno != 0 {
		return false
	}

	rootKey, ok := u.enc.Decrypt(masterKey, rootCipher)
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
	u.privateKey = privKey
	return true
}

func (u *User) changePassword(newPassword string) bool {
	if u.username == "" || u.password == "" || newPassword == "" {
		return false
	}
	p := defaultParams()
	salt := make([]byte, p.saltLength)
	_, err := rand.Read(salt)
	if err != nil {
		return false
	}
	newMasterKey := argon2.IDKey([]byte(newPassword), salt, p.iterations, p.memory, p.parallelism, p.keyLength)
	hashMaster := sha256.New()
	_, err = hashMaster.Write(newMasterKey)
	if err != nil {
		return false
	}
	hashMasterKey := hashMaster.Sum(nil)

	privKeyBytes := x509.MarshalPKCS1PrivateKey(u.privateKey)

	rootCipher, ok := u.enc.Encrypt(newMasterKey, u.rootKey)
	if ok != nil {
		return false
	}
	privCipher, ok := u.enc.Encrypt(newMasterKey, privKeyBytes)
	if ok != nil {
		return false
	}

	errno := u.m.ChangePassword(u.username, hashMasterKey, salt, rootCipher, privCipher)
	if errno != 0 {
		return false
	}

	u.password = newPassword
	u.masterKey = newMasterKey
	return true
}

func (u *User) shareDir(dir, username string) bool {
	info, err := os.Stat(dir)
	if err != nil {
		fmt.Println("Error getting file info:", err)
		return false
	}

	stat, ok := info.Sys().(*syscall.Stat_t)
	if !ok {
		fmt.Println("Not a syscall.Stat_t")
		return false
	}

	if stat == nil {
		fmt.Println("Stat is nil")
		return false
	}

	// directory ?
	if !info.IsDir() {
		return false
	}
	inode := stat.Ino

	var userId uint32
	err = u.m.GetUserId(username, &userId)
	if err != nil {
		return false
	}

	var keys [][]byte
	errno := u.m.GetPathKey(meta.Ino(inode), &keys)
	if errno != 0 {
		return false
	}

	// start at the root of the path
	key := u.rootKey
	for i := len(keys) - 1; i >= 0; i-- {
		key, err = u.enc.Decrypt(key, keys[i])
		if err != nil {
			return false
		}
	}

	name := []byte(info.Name())
	nameCipher, err := u.enc.Encrypt(key, name)
	if err != nil {
		return false
	}
	fmt.Println("Sharing", string(name), "with", username, "key:", key)

	var pubKeyBytes []byte
	err = u.m.GetUserPublicKey(username, &pubKeyBytes)
	if err != nil {
		return false
	}

	pubKey, err := x509.ParsePKCS1PublicKey(pubKeyBytes)
	if err != nil {
		return false
	}
	key, err = u.enc.EncryptRSA(pubKey, key)
	if err != nil {
		return false
	}

	errno = u.m.ShareDir(userId, meta.Ino(inode), nameCipher, key)
	return errno == 0
}
