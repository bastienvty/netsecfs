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

type MyRootNode struct {
	MyNode
}

type MyNode struct {
	fs.Inode
}

func (n *MyNode) isRoot() bool {
	_, parent := n.Parent()
	return parent == nil
}

var _ = (fs.InodeEmbedder)((*MyNode)(nil))
var _ = (fs.NodeGetattrer)((*MyNode)(nil))
var _ = (fs.NodeLookuper)((*MyNode)(nil))
var _ = (fs.NodeMknoder)((*MyNode)(nil))

func (n *MyNode) Lookup(ctx context.Context, name string, out *fuse.EntryOut) (ch *fs.Inode, errno syscall.Errno) {
	/*ops := MyNode{}
	out.Mode = 0755
	out.Size = 42
	fmt.Println("Lookup ", name)
	return n.NewInode(ctx, &ops, fs.StableAttr{Mode: syscall.S_IFREG}), 0*/
	if child := n.GetChild(name); child != nil {
		return child, fs.OK
	}
	return nil, syscall.ENOENT
}

func (n *MyNode) Getattr(ctx context.Context, f fs.FileHandle, out *fuse.AttrOut) (errno syscall.Errno) {
	if f != nil {
		return f.(fs.FileGetattrer).Getattr(ctx, out)
	}
	out.Mode = n.Mode()
	out.Size = 456
	return fs.OK
}

func (n *MyNode) Access(ctx context.Context, mode uint32) syscall.Errno {
	return 0
}

func (n *MyNode) Open(ctx context.Context, flags uint32) (fh fs.FileHandle, fuseFlags uint32, errno syscall.Errno) {
	return nil, 0, 0
}

func (n *MyNode) Statfs(ctx context.Context, out *fuse.StatfsOut) syscall.Errno {
	return 0
}

// Mknod
func (n *MyNode) Mknod(ctx context.Context, name string, mode, rdev uint32, out *fuse.EntryOut) (inode *fs.Inode, errno syscall.Errno) {
	fmt.Println("Mknod ", name)
	return
}
