package fs

import (
	"context"
	"crypto/rand"
	"syscall"

	"github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"
)

type File struct {
	n *Node
}

var _ fs.FileHandle = (*File)(nil)

var _ = (fs.FileReader)((*File)(nil))

var _ = (fs.FileWriter)((*File)(nil))

var _ = (fs.FileFlusher)((*File)(nil))
var _ = (fs.FileReleaser)((*File)(nil))
var _ = (fs.FileFsyncer)((*File)(nil))

func (f *File) Read(ctx context.Context, dest []byte, off int64) (fuse.ReadResult, syscall.Errno) {
	ino := f.n.StableAttr().Ino
	var keyCipher []byte
	dataCipher, error := f.n.obj.Get(ino, off, &keyCipher)
	if error != nil {
		return nil, syscall.EIO
	}
	key, ok := f.n.enc.Decrypt(f.n.key, keyCipher)
	if ok != nil {
		return nil, syscall.EIO
	}
	data, ok := f.n.enc.Decrypt(key, dataCipher)
	if ok != nil {
		return nil, syscall.EIO
	}
	return fuse.ReadResultData(data), 0
}

func (f *File) Write(ctx context.Context, data []byte, off int64) (written uint32, errno syscall.Errno) {
	ino := f.n.StableAttr().Ino
	err := f.n.meta.Write(ctx, ino, data, off)
	if err != 0 {
		return 0, err
	}
	size := int64(len(data))
	key := f.n.key
	contentKey := make([]byte, 32)
	_, ok := rand.Read(contentKey)
	if ok != nil {
		return 0, syscall.EIO
	}
	contentKeyCipher, ok := f.n.enc.Encrypt(key, contentKey)
	if ok != nil {
		return 0, syscall.EIO
	}
	dataCipher, ok := f.n.enc.Encrypt(contentKey, data)
	if ok != nil {
		return 0, syscall.EIO
	}
	error := f.n.obj.Put(ino, contentKeyCipher, dataCipher, size)
	if error != nil {
		return 0, syscall.EIO
	}
	return uint32(len(data)), 0
}

func (f *File) Flush(ctx context.Context) syscall.Errno {
	return 0
}

func (f *File) Release(ctx context.Context) syscall.Errno {
	return 0
}

func (f *File) Fsync(ctx context.Context, flags uint32) syscall.Errno {
	return 0
}
