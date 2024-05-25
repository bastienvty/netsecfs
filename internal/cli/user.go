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

func (u *User) createUser() bool {
	if u.username == "" || u.password == "" {
		fmt.Println("Username or password is empty.")
		return false
	}
	if u.m.CheckUser(u.username) != nil {
		fmt.Printf("User %s already exists.\n", u.username)
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

	err = u.m.CreateUser(u.username, hashMasterKey, salt, rootCipher, privCipher, pubKeyBytes)
	if err != nil {
		return false
	}

	u.masterKey = masterKey
	u.rootKey = rootKey
	u.privateKey = privKey
	return true
}

func (u *User) verifyUser() bool {
	if u.username == "" || u.password == "" {
		fmt.Println("Username or password is empty.")
		return false
	}
	p := defaultParams()
	var salt []byte
	err := u.m.GetSalt(u.username, &salt)
	if err != nil {
		return false
	}
	masterKey := argon2.IDKey([]byte(u.password), salt, p.iterations, p.memory, p.parallelism, p.keyLength)

	hashMaster := sha256.New()
	_, err = hashMaster.Write(masterKey)
	if err != nil {
		return false
	}
	hashMasterKey := hashMaster.Sum(nil)
	var rootCipher, privCipher []byte
	err = u.m.VerifyUser(u.username, hashMasterKey, &rootCipher, &privCipher)
	if err != nil {
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
		fmt.Println("Username or password is empty.")
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

	err = u.m.ChangePassword(u.username, hashMasterKey, salt, rootCipher, privCipher)
	if err != nil {
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
		return false
	}

	if stat == nil {
		return false
	}

	if !info.IsDir() {
		fmt.Printf("%s is not a directory.\n", dir)
		return false
	}
	inode := stat.Ino

	var userId uint32
	err = u.m.GetUserId(username, &userId)
	if err != nil {
		fmt.Printf("No such user found: %s\n", username)
		return false
	}

	var keys [][]byte
	err = u.m.GetPathKey(meta.Ino(inode), &keys)
	if err != nil {
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

	nameSign, err := u.enc.Sign(u.privateKey, name)
	if err != nil {
		return false
	}

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

	err = u.m.ShareDir(userId, meta.Ino(inode), nameCipher, key, nameSign)
	return err == nil
}

func (u *User) unshareDir(dir, username string) bool {
	info, err := os.Stat(dir)
	if err != nil {
		fmt.Println("Error getting file info:", err)
		return false
	}

	stat, ok := info.Sys().(*syscall.Stat_t)
	if !ok {
		return false
	}

	if stat == nil {
		return false
	}

	if !info.IsDir() {
		fmt.Printf("%s is not a directory.\n", dir)
		return false
	}
	inode := stat.Ino

	var userId uint32
	err = u.m.GetUserId(username, &userId)
	if err != nil {
		fmt.Printf("No such user found: %s\n", username)
		return false
	}

	var sign []byte
	err = u.m.VerifyShare(userId, meta.Ino(inode), &sign)
	if err != nil {
		fmt.Println("No corresponding share found.")
		return false
	}

	pubKey := &u.privateKey.PublicKey
	err = u.enc.VerifySign(pubKey, []byte(info.Name()), sign)
	if err != nil {
		fmt.Println("You are not the owner of this directory. You cannot unshare it.")
		return false
	}

	err = u.m.UnshareDir(userId, meta.Ino(inode))
	return err == nil
}
