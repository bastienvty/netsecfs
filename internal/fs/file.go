package fs

import (
	"os"
	"sync"
	"syscall"

	"github.com/bastienvty/netsecfs/internal/db/meta"
	"github.com/hanwen/go-fuse/v2/fs"
)

// https://github.com/aegistudio/enigma/blob/master/cmd/enigma/fuse_unix.go
// https://github.com/pachyderm/pachyderm/blob/master/src/server/pfs/fuse/files.go
// https://github.com/rclone/rclone/blob/b2f6aac754c5d46c66758db46ecb89aa85c3c113/cmd/mount2/file.go
// https://github.com/materials-commons/hydra/blob/main/pkg/mcfs/fs/mcfs/base_file_handle.go
// juicefs
// nanafs
// gocryptfs

type FileHandle struct {
	ino meta.Ino
	mu  sync.Mutex
	fd  *os.File
}

var _ fs.FileHandle = (*FileHandle)(nil)

// var _ = (fs.FileGetattrer)((*File)(nil))
// var _ = (fs.FileReader)((*File)(nil))
// var _ = (fs.FileWriter)((*File)(nil))
// var _ = (fs.FileFlusher)((*File)(nil))
// var _ = (fs.FileReleaser)((*File)(nil))
// var _ = (fs.FileFsyncer)((*File)(nil))

func newFileHandle(ino meta.Ino, name string) (fh *FileHandle, errno syscall.Errno) {
	st := &syscall.Stat_t{}
	if err := syscall.Fstat(int(ino), st); err != nil {
		errno = fs.ToErrno(err)
		return
	}

	osFile := os.NewFile(uintptr(ino), name)

	fh = &FileHandle{
		ino: ino,
		fd:  osFile,
	}

	return fh, 0
}

/*func (f *NSFile) Getattr(ctx context.Context, out *fuse.AttrOut) syscall.Errno {
	return 0
}

func (f *File) Read(ctx context.Context, dest []byte, off int64) (fuse.ReadResult, syscall.Errno) {
	// s := iofs.StatFS().FS
	return nil, 0
}*/

/*func (f *NSFile) Write(ctx context.Context, data []byte, off int64) (written uint32, errno syscall.Errno) {
	return 0, 0
}

func (f *NSFile) Flush(ctx context.Context) syscall.Errno {
	return 0
}

func (f *NSFile) Release(ctx context.Context) syscall.Errno {
	return 0
}

func (f *NSFile) Fsync(ctx context.Context, flags uint32) syscall.Errno {
	return 0
}*/
