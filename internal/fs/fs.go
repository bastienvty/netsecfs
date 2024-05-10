package fs

import (
	"github.com/bastienvty/netsecfs/internal/config"
)

const (
	fsName = "netsecfs"
)

type FS struct {
	Path string
	Name string

	cfg config.FUSE

	debug bool
}

func NewFS(cfg config.FUSE) *FS {
	return &FS{
		cfg: cfg,
	}
}
