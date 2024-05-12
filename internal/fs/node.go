package fs

import (
	"context"
	"fmt"
	"syscall"

	"github.com/bastienvty/netsecfs/internal/db/meta"
	"github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"
)

// DO NOT USE NODEFS OR PATHFS. THEY ARE DEPRECATED.
// USE THE FS PACKAGE INSTEAD: https://github.com/hanwen/go-fuse/blob/master/fs/api.go

// https://github.com/materials-commons/hydra/blob/main/pkg/mcfs/fs/mcbridgefs/node.go
//

const (
	rootID  = 1
	maxName = meta.MaxName
	/*maxSymlink  = meta.MaxSymlink
	maxFileSize = meta.ChunkSize << 31*/
)

type Node struct {
	fs.Inode
}

func (n *Node) isRoot() bool {
	_, parent := n.Parent()
	return parent == nil
}

var _ = (fs.InodeEmbedder)((*Node)(nil))
var _ = (fs.NodeLookuper)((*Node)(nil))
var _ = (fs.NodeGetattrer)((*Node)(nil))

/*var _ = (fs.NodeStatfser)((*Node)(nil))

// var _ = (fs.NodeOpener)((*Node)(nil))
var _ = (fs.NodeCreater)((*Node)(nil))
var _ = (fs.NodeRenamer)((*Node)(nil))

var _ = (fs.NodeAccesser)((*Node)(nil))
var _ = (fs.NodeOpendirer)((*Node)(nil))
var _ = (fs.NodeReaddirer)((*Node)(nil))
var _ = (fs.NodeMkdirer)((*Node)(nil))
var _ = (fs.NodeRmdirer)((*Node)(nil))

var _ = (fs.NodeUnlinker)((*Node)(nil)) // vim
var _ = (fs.NodeFsyncer)((*Node)(nil))  // vim*/

func (n *Node) Lookup(ctx context.Context, name string, out *fuse.EntryOut) (*fs.Inode, syscall.Errno) {
	if len(name) > maxName {
		return nil, syscall.ENAMETOOLONG
	}
	fmt.Println("Lookup node", n, "with ino", n.Inode.StableAttr().Ino)
	fmt.Println("Lookup name", name)
	ops := Node{}
	out.Mode = 0755
	out.Size = 42
	return n.NewInode(ctx, &ops, fs.StableAttr{Mode: syscall.S_IFREG}), 0
}

func (n *Node) Getattr(ctx context.Context, f fs.FileHandle, out *fuse.AttrOut) (errno syscall.Errno) {
	if f != nil {
		return f.(fs.FileGetattrer).Getattr(ctx, out)
	}
	out.Mode = n.Mode()
	out.Size = 456
	return fs.OK
}

/*func (n *Node) Statfs(ctx context.Context, out *fuse.StatfsOut) syscall.Errno {
	return 0
}

func (n *Node) Access(ctx context.Context, mask uint32) syscall.Errno {
	return 0
}*/

/*func (n *Node) Open(ctx context.Context, flags uint32) (fh fs.FileHandle, fuseFlags uint32, errno syscall.Errno) {
	return nil, 0, 0
}*/

/*func (n *Node) Create(ctx context.Context, name string, flags uint32, mode uint32, out *fuse.EntryOut) (node *fs.Inode, fh fs.FileHandle, fuseFlags uint32, errno syscall.Errno) {
	return nil, nil, 0, 0
}

func (n *Node) Rename(ctx context.Context, name string, newParent fs.InodeEmbedder, newName string, flags uint32) syscall.Errno {
	return 0
}

func (n *Node) Opendir(ctx context.Context) syscall.Errno {
	return 0
}

func (n *Node) Readdir(ctx context.Context) (fs.DirStream, syscall.Errno) {
	return nil, 0
}

func (n *Node) Mkdir(ctx context.Context, name string, mode uint32, out *fuse.EntryOut) (*fs.Inode, syscall.Errno) {
	return nil, 0
}

func (n *Node) Rmdir(ctx context.Context, name string) syscall.Errno {
	return 0
}

func (n *Node) Unlink(ctx context.Context, name string) syscall.Errno {
	return 0
}

func (n *Node) Fsync(ctx context.Context, f fs.FileHandle, flags uint32) syscall.Errno {
	return 0
}*/
