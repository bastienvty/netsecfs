package fs

import (
	"context"
	"fmt"
	"syscall"

	"github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"
)

// DO NOT USE NODEFS OR PATHFS. THEY ARE DEPRECATED.
// USE THE FS PACKAGE INSTEAD: https://github.com/hanwen/go-fuse/blob/master/fs/api.go

// https://github.com/materials-commons/hydra/blob/main/pkg/mcfs/fs/mcbridgefs/node.go
//

type NetSRootNode struct {
	NetSNode
	Path string
	Dev  uint64
}

type NetSNode struct {
	fs.Inode
}

func NewNetSNode(data []byte) *NetSNode {
	return &NetSNode{}
}

func (n *NetSNode) isRoot() bool {
	_, parent := n.Parent()
	return parent == nil
}

var _ = (fs.InodeEmbedder)((*NetSNode)(nil))

var _ = (fs.NodeLookuper)((*NetSNode)(nil))
var _ = (fs.NodeGetattrer)((*NetSNode)(nil))
var _ = (fs.NodeStatfser)((*NetSNode)(nil))

// var _ = (fs.NodeOpener)((*NetSNode)(nil))
var _ = (fs.NodeCreater)((*NetSNode)(nil))
var _ = (fs.NodeRenamer)((*NetSNode)(nil))

var _ = (fs.NodeAccesser)((*NetSNode)(nil))
var _ = (fs.NodeOpendirer)((*NetSNode)(nil))
var _ = (fs.NodeReaddirer)((*NetSNode)(nil))
var _ = (fs.NodeMkdirer)((*NetSNode)(nil))
var _ = (fs.NodeRmdirer)((*NetSNode)(nil))

var _ = (fs.NodeUnlinker)((*NetSNode)(nil)) // vim
var _ = (fs.NodeFsyncer)((*NetSNode)(nil))  // vim

func (n *NetSNode) Lookup(ctx context.Context, name string, out *fuse.EntryOut) (*fs.Inode, syscall.Errno) {
	ops := NetSNode{}
	out.Mode = 0755
	out.Size = 42
	fmt.Println("Lookup ", name)
	return n.NewInode(ctx, &ops, fs.StableAttr{Mode: syscall.S_IFREG}), 0
}

func (n *NetSNode) Getattr(ctx context.Context, f fs.FileHandle, out *fuse.AttrOut) (errno syscall.Errno) {
	if f != nil {
		return f.(fs.FileGetattrer).Getattr(ctx, out)
	}
	out.Mode = n.Mode()
	out.Size = 456
	return fs.OK
}

func (n *NetSNode) Statfs(ctx context.Context, out *fuse.StatfsOut) syscall.Errno {
	return 0
}

func (n *NetSNode) Access(ctx context.Context, mask uint32) syscall.Errno {
	return 0
}

/*func (n *NetSNode) Open(ctx context.Context, flags uint32) (fh fs.FileHandle, fuseFlags uint32, errno syscall.Errno) {
	return nil, 0, 0
}*/

func (n *NetSNode) Create(ctx context.Context, name string, flags uint32, mode uint32, out *fuse.EntryOut) (node *fs.Inode, fh fs.FileHandle, fuseFlags uint32, errno syscall.Errno) {
	return nil, nil, 0, 0
}

func (n *NetSNode) Rename(ctx context.Context, name string, newParent fs.InodeEmbedder, newName string, flags uint32) syscall.Errno {
	return 0
}

func (n *NetSNode) Opendir(ctx context.Context) syscall.Errno {
	return 0
}

func (n *NetSNode) Readdir(ctx context.Context) (fs.DirStream, syscall.Errno) {
	return nil, 0
}

func (n *NetSNode) Mkdir(ctx context.Context, name string, mode uint32, out *fuse.EntryOut) (*fs.Inode, syscall.Errno) {
	return nil, 0
}

func (n *NetSNode) Rmdir(ctx context.Context, name string) syscall.Errno {
	return 0
}

func (n *NetSNode) Unlink(ctx context.Context, name string) syscall.Errno {
	return 0
}

func (n *NetSNode) Fsync(ctx context.Context, f fs.FileHandle, flags uint32) syscall.Errno {
	return 0
}
