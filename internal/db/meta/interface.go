/*
 * JuiceFS, Copyright 2020 Juicedata, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

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
	RenameNoReplace = 1 << iota
	RenameExchange
	RenameWhiteout
	_renameReserved1
	_renameReserved2
	RenameRestore // internal
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

// Type of control messages
const CPROGRESS = 0xFE // 16 bytes: progress increment
const CDATA = 0xFF     // 4 bytes: data length

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

	// StatFS returns summary statistics of a volume (no need here as it is handled in node, just for reference)
	// StatFS(ctx context.Context, ino Ino, totalspace, availspace, iused, iavail *uint64) syscall.Errno
	// Access checks the access permission on given inode.
	// doAccess(ctx context.Context, inode Ino, modemask uint8, attr *Attr) syscall.Errno
	// Lookup returns the inode and attributes for the given entry in a directory.
	Lookup(ctx context.Context, parent Ino, name string, inode *Ino, attr *Attr) syscall.Errno
	// GetAttr returns the attributes for given node.
	GetAttr(ctx context.Context, inode Ino, attr *Attr) syscall.Errno
	// SetAttr updates the attributes for given node.
	SetAttr(ctx context.Context, inode Ino, in *fuse.SetAttrIn, attr *Attr) syscall.Errno
	// doMknod(ctx context.Context, parent Ino, name string, _type uint8, mode uint16, cumask uint16, rdev uint32, path string, inode *Ino, attr *Attr) syscall.Errno
	// Mkdir creates a sub-directory with given name and mode.
	// doMkdir(ctx context.Context, parent Ino, name string, mode uint16, cumask uint16, copysgid uint8, inode *Ino, attr *Attr) syscall.Errno
	// Unlink removes a file entry from a directory.
	// The file will be deleted if it's not linked by any entries and not open by any sessions.
	Unlink(ctx context.Context, parent Ino, name string) syscall.Errno
	// Rmdir removes an empty sub-directory.
	Rmdir(ctx context.Context, parent Ino, name string) syscall.Errno
	// Rename move an entry from a source directory to another with given name.
	// The targeted entry will be overwrited if it's a file or empty directory.
	// For Hadoop, the target should not be overwritten.
	// doRename(ctx context.Context, parentSrc Ino, nameSrc string, parentDst Ino, nameDst string, flags uint32, inode *Ino, attr *Attr) syscall.Errno
	// Readdir returns all entries for given directory, which include attributes if plus is true.
	Readdir(ctx context.Context, inode Ino, wantattr uint8, entries *[]*Entry) syscall.Errno
	// Create creates a file in a directory with given name.
	// Create(ctx context.Context, parent Ino, name string, mode uint16, cumask uint16, flags uint32, inode *Ino, attr *Attr) syscall.Errno
	Mknod(ctx context.Context, parent Ino, name string, _type uint8, mode uint32, inode *Ino, attr *Attr) syscall.Errno
	// Open checks permission on a node and track it as open.
	// doOpen(ctx context.Context, inode Ino, flags uint32, attr *Attr) syscall.Errno
	// Close a file.
	// doClose(ctx context.Context, inode Ino) syscall.Errno
	// Read returns the list of slices on the given chunk.
	// doRead(ctx context.Context, inode Ino, indx uint32, slices *[]Slice) syscall.Errno
	// Write put a slice of data on top of the given chunk.
	// doWrite(ctx context.Context, inode Ino, indx uint32, off uint32, slice Slice, mtime time.Time) syscall.Errno
}

func RegisterMeta(addr string) Meta {
	m, err := newSQLMeta("sqlite3", addr)
	if err != nil {
		logger.Fatalf("unable to register client: %s", err)
	}
	return m
}
