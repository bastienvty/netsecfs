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
	fileBlockSize = 1 << 12          // 4k
	maxSize       = 1125899906842624 // 1TB
)

type RootNode struct {
	fs.Inode
}

type Node struct {
	fs.Inode
	meta meta.Meta
}

func NewNode(meta meta.Meta) *Node {
	return &Node{meta: meta}
}

func (n *Node) isRoot() bool {
	_, parent := n.Parent()
	return parent == nil
}

var _ = (fs.InodeEmbedder)((*Node)(nil))
var _ = (fs.NodeLookuper)((*Node)(nil))
var _ = (fs.NodeGetattrer)((*Node)(nil))
var _ = (fs.NodeStatfser)((*Node)(nil))

// var _ = (fs.NodeOpener)((*Node)(nil))
/*var _ = (fs.NodeCreater)((*Node)(nil))
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
	var err syscall.Errno
	var attr = &meta.Attr{}
	var inode *meta.Ino
	ino := meta.Ino(n.StableAttr().Ino)
	err = n.meta.Lookup(ctx, ino, name, inode, attr)
	if err != 0 {
		return nil, err
	}
	fmt.Println("LOOKUP", ino, name, inode, attr)
	ops := Node{}
	out.Mode = 0755
	out.Size = 42
	return n.NewInode(ctx, &ops, fs.StableAttr{Mode: syscall.S_IFREG}), 0
}

func attrToStat(inode meta.Ino, attr *meta.Attr, out *fuse.Attr) {
	out.Ino = uint64(inode)
	out.Uid = attr.Uid
	out.Gid = attr.Gid
	out.Mode = attr.SMode()
	out.Nlink = attr.Nlink
	out.Atime = uint64(attr.Atime)
	out.Atimensec = attr.Atimensec
	out.Mtime = uint64(attr.Mtime)
	out.Mtimensec = attr.Mtimensec
	out.Ctime = uint64(attr.Ctime)
	out.Ctimensec = attr.Ctimensec

	var size, blocks uint64
	switch attr.Typ {
	case meta.TypeDirectory:
		fallthrough
	case meta.TypeFile:
		size = attr.Length
		blocks = (size + 511) / 512
	}
	out.Size = size
	out.Blocks = blocks
	out.Blksize = 4096
}

func (n *Node) Getattr(ctx context.Context, f fs.FileHandle, out *fuse.AttrOut) (errno syscall.Errno) {
	if f != nil {
		return f.(fs.FileGetattrer).Getattr(ctx, out)
	}
	var err syscall.Errno
	var attr = &meta.Attr{}
	ino := meta.Ino(n.StableAttr().Ino)
	err = n.meta.GetAttr(ctx, ino, attr)
	if err == 0 {
		entry := &meta.Entry{Inode: ino, Attr: attr}
		attrToStat(entry.Inode, entry.Attr, &out.Attr)
	}
	return err
}

/*func fsInfo2StatFs(out *fuse.StatfsOut) {
	out.Blocks = maxSize / fileBlockSize
	out.Bfree = (maxSize - info.UsageSize) / fileBlockSize
	out.Bavail = out.Bfree
	out.Files = info.AvailInodes
	out.Ffree = info.AvailInodes - info.Objects
	out.Bsize = uint32(fileBlockSize)
	out.NameLen = maxName
}*/

func (n *Node) Statfs(ctx context.Context, out *fuse.StatfsOut) syscall.Errno {
	out.Blocks = uint64(maxSize) / fileBlockSize    // Total data blocks in file system.
	out.Bfree = uint64(maxSize-1e9) / fileBlockSize // Free blocks in file system.
	out.Bavail = out.Bfree                          // Free blocks in file system if you're not root.
	out.Files = 1e9                                 // Total files in file system.
	out.Ffree = 1e9                                 // Free files in file system.
	out.Bsize = fileBlockSize                       // Block size
	out.NameLen = 255                               // Maximum file name length?
	out.Frsize = fileBlockSize                      // Fragment size, smallest addressable data size in the file system.
	fmt.Println("STATFS", out)
	return 0
}

/*func (n *Node) Access(ctx context.Context, mask uint32) syscall.Errno {
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
