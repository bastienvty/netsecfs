package fs

// https://github.com/aegistudio/enigma/blob/master/cmd/enigma/fuse_unix.go
// https://github.com/pachyderm/pachyderm/blob/master/src/server/pfs/fuse/files.go
// https://github.com/rclone/rclone/blob/b2f6aac754c5d46c66758db46ecb89aa85c3c113/cmd/mount2/file.go
// https://github.com/materials-commons/hydra/blob/main/pkg/mcfs/fs/mcfs/base_file_handle.go
// juicefs
// nanafs
// gocryptfs

/*type NSFile struct {
	NetSNode
	mu sync.Mutex
	fd int
}

var _ = (fs.FileGetattrer)((*NSFile)(nil))
var _ = (fs.FileReader)((*NSFile)(nil))
var _ = (fs.FileWriter)((*NSFile)(nil))
var _ = (fs.FileFlusher)((*NSFile)(nil))
var _ = (fs.FileReleaser)((*NSFile)(nil))
var _ = (fs.FileFsyncer)((*NSFile)(nil))

func NewNSFile(data []byte) *NSFile {
	return &NSFile{}
}

func (f *NSFile) Getattr(ctx context.Context, out *fuse.AttrOut) syscall.Errno {
	return 0
}

func (f *NSFile) Read(ctx context.Context, dest []byte, off int64) (fuse.ReadResult, syscall.Errno) {
	// s := iofs.StatFS().FS
	return nil, 0
}

func (f *NSFile) Write(ctx context.Context, data []byte, off int64) (written uint32, errno syscall.Errno) {
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
