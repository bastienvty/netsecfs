package meta

import (
	"context"
	"strconv"
	"syscall"
	"time"

	"github.com/hanwen/go-fuse/v2/fuse"
)

const (
	TypeFile      = 1 // type for regular file
	TypeDirectory = 2 // type for directory
)

const (
	// SetAttrMode is a mask to update a attribute of node
	SetAttrMode = 1 << iota
	SetAttrUID
	SetAttrGID
	SetAttrSize
	SetAttrAtime
	SetAttrMtime
	SetAttrCtime
	SetAttrAtimeNow
	SetAttrMtimeNow
)

const MaxName = 255

type Ino uint64

const RootInode Ino = 1
const SharedInode Ino = 2
const SkipDirMtime time.Duration = 100 * time.Millisecond

func (i Ino) String() string {
	return strconv.FormatUint(uint64(i), 10)
}

func (i Ino) IsValid() bool {
	return i >= RootInode
}

func (i Ino) IsNormal() bool {
	return i >= RootInode
}

// Attr represents attributes of a node.
type Attr struct {
	Typ       uint8  // type of a node
	Mode      uint16 // permission mode
	Rdev      uint32 // device number
	Atime     int64  // last access time
	Mtime     int64  // last modified time
	Ctime     int64  // last change time for meta
	Atimensec uint32 // nanosecond part of atime
	Mtimensec uint32 // nanosecond part of mtime
	Ctimensec uint32 // nanosecond part of ctime
	Nlink     uint32 // number of links (sub-directories or hardlinks)
	Length    uint64 // length of regular file

	Parent Ino  // inode of parent; 0 means tracked by parentKey (for hardlinks)
	Full   bool // the attributes are completed or not
}

func typeToStatType(_type uint8) uint32 {
	switch _type & 0x7F {
	case TypeDirectory:
		return syscall.S_IFDIR
	case TypeFile:
		return syscall.S_IFREG
	default:
		panic(_type)
	}
}

func typeToString(_type uint8) string {
	switch _type {
	case TypeFile:
		return "regular"
	case TypeDirectory:
		return "directory"
	default:
		return "unknown"
	}
}

func typeFromString(s string) uint8 {
	switch s {
	case "regular":
		return TypeFile
	case "directory":
		return TypeDirectory
	default:
		panic(s)
	}
}

// SMode is the file mode including type and unix permission.
func (a Attr) SMode() uint32 {
	return typeToStatType(a.Typ) | uint32(a.Mode)
}

// Entry is an entry inside a directory.
type Entry struct {
	Inode Ino
	Name  []byte
	Key   []byte
	Attr  *Attr
}

// Meta is a interface for a meta service for file system.
type Meta interface {
	// Name of database
	Name() string
	// Init is used to initialize a meta service.
	Init(format *Format) error
	// Shutdown close current database connections.
	Shutdown()
	Load() (*Format, error)
	GetNextInode(ctx context.Context, lastIno *Ino) error
	GetUserId(username string, uid *uint32) error
	GetUserPublicKey(username string, pubKey *[]byte) error

	// Lookup returns the inode and attributes for the given entry in a directory.
	Lookup(ctx context.Context, userId uint32, parent, inode Ino, attr *Attr) syscall.Errno
	// GetAttr returns the attributes for given node.
	GetAttr(ctx context.Context, inode Ino, attr *Attr) syscall.Errno
	// SetAttr updates the attributes for given node.
	SetAttr(ctx context.Context, inode Ino, in *fuse.SetAttrIn, attr *Attr) syscall.Errno
	// Unlink removes a file entry from a directory.
	// The file will be deleted if it's not linked by any entries and not open by any sessions.
	Unlink(ctx context.Context, parent, inode Ino) syscall.Errno
	// Rmdir removes an empty sub-directory.
	Rmdir(ctx context.Context, parent, inode Ino) syscall.Errno
	// Readdir returns all entries for given directory, which include attributes if plus is true.
	Readdir(ctx context.Context, inode Ino, userId uint32, entries *[]*Entry) syscall.Errno
	Mknod(ctx context.Context, parent Ino, _type uint8, mode, id uint32, inode *Ino, name, key []byte, attr *Attr) syscall.Errno
	// Write put a slice of data on top of the given chunk.
	Write(ctx context.Context, inode uint64, data []byte, off int64) syscall.Errno
	GetKey(ctx context.Context, inode Ino, key *[]byte) syscall.Errno
	GetSharedKey(ctx context.Context, userdId uint32, inode Ino, key *[]byte) syscall.Errno

	CheckUser(username string) error
	CreateUser(username string, password, salt, rootKey, privKey, pubKey []byte) error
	VerifyUser(username string, password []byte, rootKey, privKey *[]byte) error
	GetSalt(username string, salt *[]byte) error
	ChangePassword(username string, password, salt, rootKey, privKey []byte) error
	ShareDir(user uint32, inode Ino, name, key []byte) error
	UnshareDir(user uint32, inode Ino) error
	GetPathKey(inode Ino, keys *[][]byte) error
}

func RegisterMeta(addr string) Meta {
	m, err := newSQLMeta("sqlite3", addr)
	if err != nil {
		logger.Fatalf("unable to register client: %s", err)
	}
	return m
}
