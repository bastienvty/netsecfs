package fs

import (
	"context"
	"fmt"
	"syscall"

	"github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"
)

// https://github.com/aegistudio/enigma/blob/master/cmd/enigma/fuse_unix.go
// https://github.com/pachyderm/pachyderm/blob/master/src/server/pfs/fuse/files.go
// https://github.com/rclone/rclone/blob/b2f6aac754c5d46c66758db46ecb89aa85c3c113/cmd/mount2/file.go
// https://github.com/materials-commons/hydra/blob/main/pkg/mcfs/fs/mcfs/base_file_handle.go
// juicefs
// nanafs
// gocryptfs

type FileHandle struct {
	n *Node
}

var _ fs.FileHandle = (*FileHandle)(nil)

// var _ = (fs.FileGetattrer)((*FileHandle)(nil))
// var _ = (fs.FileSetattrer)((*FileHandle)(nil))
var _ = (fs.FileReader)((*FileHandle)(nil))

var _ = (fs.FileWriter)((*FileHandle)(nil))

var _ = (fs.FileFlusher)((*FileHandle)(nil))
var _ = (fs.FileReleaser)((*FileHandle)(nil))
var _ = (fs.FileFsyncer)((*FileHandle)(nil))

/*func newFileHandle(meta meta.Meta, name string) (fh *FileHandle, errno syscall.Errno) {
	st := &syscall.Stat_t{}
	if err := syscall.Fstat(int(ino), st); err != nil {
		errno = fs.ToErrno(err)
		return
	}

	osFile := os.NewFile(uintptr(ino), name)

	fh = &FileHandle{}

	return fh, 0
}*/

/*func (f *NSFile) Getattr(ctx context.Context, out *fuse.AttrOut) syscall.Errno {
	return 0
}*/

func (f *FileHandle) Read(ctx context.Context, dest []byte, off int64) (fuse.ReadResult, syscall.Errno) {
	fmt.Println("READ FILE", dest, string(dest), len(dest), off)
	ino := f.n.StableAttr().Ino
	data, error := f.n.obj.Get(ino, "nil", off)
	if error != nil {
		fmt.Println("ERROR GET:", data)
		return nil, syscall.EIO
	}
	fmt.Println("READ DATA:", string(data))
	return fuse.ReadResultData(data), 0
}

func (f *FileHandle) Write(ctx context.Context, data []byte, off int64) (written uint32, errno syscall.Errno) {
	fmt.Println("DATA:", string(data))
	fmt.Println("OFF:", off)
	ino := f.n.StableAttr().Ino
	/*text := string(data)
	lines := strings.Split(text, "\n")
	if len(lines) > 2 {
		lines = lines[:len(lines)-2]
		text = strings.Join(lines, "\n") + "\n"
	}
	fmt.Println("TEXT:", text)
	newData := []byte(text)*/
	err := f.n.meta.Write(ctx, ino, data, off)
	if err != 0 {
		return 0, err
	}
	// key := uuid.New().String()
	error := f.n.obj.Put(ino, "nil", data)
	if error != nil {
		fmt.Println("ERROR PUT:", error)
		return 0, syscall.EIO
	}
	return uint32(len(data)), 0
}

func (f *FileHandle) Flush(ctx context.Context) syscall.Errno {
	fmt.Println("FLUSH FILE")
	return 0
}

func (f *FileHandle) Release(ctx context.Context) syscall.Errno {
	fmt.Println("RELEASE FILE")
	return 0
}

func (f *FileHandle) Fsync(ctx context.Context, flags uint32) syscall.Errno {
	fmt.Println("FSYNC FILE")
	return 0
}
