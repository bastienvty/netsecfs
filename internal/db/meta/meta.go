package meta

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"runtime"
	"runtime/debug"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/bastienvty/netsecfs/utils"
	"github.com/hanwen/go-fuse/v2/fuse"
	_ "github.com/mattn/go-sqlite3"
	"github.com/pkg/errors"
	"xorm.io/xorm"
	"xorm.io/xorm/names"
)

var logger = utils.GetLogger("netsecfs")

type setting struct {
	Name  string `xorm:"pk"`
	Value string `xorm:"varchar(4096) notnull"`
}

type edge struct {
	Id     int64  `xorm:"pk bigserial"`
	Parent Ino    `xorm:"unique(edge) notnull"`
	Name   []byte `xorm:"unique(edge) varbinary(255) notnull"`
	Inode  Ino    `xorm:"index notnull"`
	Type   uint8  `xorm:"notnull"`
	Key    []byte
}

type node struct {
	Inode     Ino    `xorm:"pk"`
	Type      uint8  `xorm:"notnull"`
	Mode      uint16 `xorm:"notnull"`
	Atime     int64  `xorm:"notnull"`
	Mtime     int64  `xorm:"notnull"`
	Ctime     int64  `xorm:"notnull"`
	Atimensec int16  `xorm:"notnull default 0"`
	Mtimensec int16  `xorm:"notnull default 0"`
	Ctimensec int16  `xorm:"notnull default 0"`
	Nlink     uint32 `xorm:"notnull"`
	Length    uint64 `xorm:"notnull"`
	Rdev      uint32
	Parent    Ino
}

type namedNode struct {
	node `xorm:"extends"`
	Name []byte `xorm:"varbinary(255)"`
}

type user struct {
	Id       uint32 `xorm:"pk autoincr"`
	Username string `xorm:"notnull unique"`
	Password string `xorm:"notnull"`
}

type dbMeta struct {
	sync.Mutex
	db   *xorm.Engine
	addr string
	fmt  *Format

	root Ino
}

func errno(err error) syscall.Errno {
	if err == nil {
		return 0
	}
	if eno, ok := err.(syscall.Errno); ok {
		return eno
	}
	logger.Errorf("error: %s\n%s", err, debug.Stack())
	return syscall.EIO
}

func (m *dbMeta) Name() string {
	return m.addr
}

func (m *dbMeta) Load() (*Format, error) {
	body, err := m.doLoad()
	if err == nil && len(body) == 0 {
		err = fmt.Errorf("database is not formatted, please run `netsecfs init ...` first")
	}
	if err != nil {
		return nil, err
	}
	var format = new(Format)
	if err = json.Unmarshal(body, format); err != nil {
		return nil, fmt.Errorf("json: %s", err)
	}
	m.Lock()
	m.fmt = format
	m.Unlock()
	return format, nil
}

func (m *dbMeta) doLoad() (data []byte, err error) {
	err = m.roTxn(func(ses *xorm.Session) error {
		if ok, err := ses.IsTableExist(&setting{}); err != nil {
			return err
		} else if !ok {
			return nil
		}
		s := setting{Name: "format"}
		ok, err := ses.Get(&s)
		if err == nil && ok {
			data = []byte(s.Value)
		}
		return err
	})
	return
}

func (m *dbMeta) Init(format *Format) error {
	if err := m.db.Sync2(new(setting)); err != nil {
		return fmt.Errorf("create table setting: %s", err)
	}
	if err := m.db.Sync2(new(edge)); err != nil {
		return fmt.Errorf("create table edge: %s", err)
	}
	if err := m.db.Sync2(new(node), new(user)); err != nil {
		return fmt.Errorf("create table node, user: %s", err)
	}

	var s = setting{Name: "format"}
	var ok bool
	err := m.roTxn(func(ses *xorm.Session) (err error) {
		ok, err = ses.Get(&s)
		return err
	})
	if err != nil {
		return err
	}

	if ok {
		var old Format
		err = json.Unmarshal([]byte(s.Value), &old)
		if err != nil {
			return fmt.Errorf("json: %s", err)
		}
		if err = format.update(&old); err != nil {
			return errors.Wrap(err, "update format")
		}
	}

	data, err := json.MarshalIndent(format, "", "")
	if err != nil {
		return fmt.Errorf("json: %s", err)
	}

	m.fmt = format
	now := time.Now()
	root := &node{
		Type:      TypeDirectory,
		Atime:     now.UnixNano() / 1e3,
		Mtime:     now.UnixNano() / 1e3,
		Ctime:     now.UnixNano() / 1e3,
		Atimensec: int16(now.UnixNano() % 1e3),
		Mtimensec: int16(now.UnixNano() % 1e3),
		Ctimensec: int16(now.UnixNano() % 1e3),
		Nlink:     2,
		Length:    4 << 10,
		Parent:    1,
	}
	shared := &node{
		Type:      TypeDirectory,
		Atime:     now.UnixNano() / 1e3,
		Mtime:     now.UnixNano() / 1e3,
		Ctime:     now.UnixNano() / 1e3,
		Atimensec: int16(now.UnixNano() % 1e3),
		Mtimensec: int16(now.UnixNano() % 1e3),
		Ctimensec: int16(now.UnixNano() % 1e3),
		Nlink:     2,
		Length:    4 << 10,
		Parent:    1,
	}
	return m.txn(func(s *xorm.Session) error {
		if ok {
			_, err = s.Update(&setting{"format", string(data)}, &setting{Name: "format"})
			return err
		} else {
			var set = &setting{"format", string(data)}
			if n, err := s.Insert(set); err != nil {
				return err
			} else if n == 0 {
				return fmt.Errorf("format is not inserted")
			}
		}

		root.Inode = 1
		root.Mode = 0755 // allow operations on root
		mustInsert(s, root)
		shared.Inode = 2
		shared.Mode = 0555
		return mustInsert(s, &edge{Parent: 1, Name: []byte("shared"), Inode: shared.Inode, Type: TypeDirectory}, shared)
	})
}

func (m *dbMeta) Shutdown() {
	m.db.Close()
}

func (m *dbMeta) parseAttr(n *node, attr *Attr) {
	if attr == nil || n == nil {
		return
	}
	attr.Typ = n.Type
	attr.Mode = n.Mode
	attr.Atime = n.Atime / 1e6
	attr.Atimensec = uint32(n.Atime%1e6*1000) + uint32(n.Atimensec)
	attr.Mtime = n.Mtime / 1e6
	attr.Mtimensec = uint32(n.Mtime%1e6*1000) + uint32(n.Mtimensec)
	attr.Ctime = n.Ctime / 1e6
	attr.Ctimensec = uint32(n.Ctime%1e6*1000) + uint32(n.Ctimensec)
	attr.Nlink = n.Nlink
	attr.Length = n.Length
	attr.Rdev = n.Rdev
	attr.Parent = n.Parent
	attr.Full = true
}

func (m *dbMeta) parseNode(attr *Attr, n *node) {
	if attr == nil || n == nil {
		return
	}
	n.Type = attr.Typ
	n.Mode = attr.Mode
	n.Atime = attr.Atime*1e6 + int64(attr.Atimensec)/1000
	n.Mtime = attr.Mtime*1e6 + int64(attr.Mtimensec)/1000
	n.Ctime = attr.Ctime*1e6 + int64(attr.Ctimensec)/1000
	n.Atimensec = int16(attr.Atimensec % 1000)
	n.Mtimensec = int16(attr.Mtimensec % 1000)
	n.Ctimensec = int16(attr.Ctimensec % 1000)
	n.Nlink = attr.Nlink
	n.Length = attr.Length
	n.Rdev = attr.Rdev
	n.Parent = attr.Parent
}

func mustInsert(s *xorm.Session, beans ...interface{}) error {
	for start, end, size := 0, 0, len(beans); end < size; start = end {
		end = start + 200
		if end > size {
			end = size
		}
		if n, err := s.Insert(beans[start:end]...); err != nil {
			return err
		} else if d := end - start - int(n); d > 0 {
			return fmt.Errorf("%d records not inserted: %+v", d, beans[start:end])
		}
	}
	return nil
}

var errBusy error

func (m *dbMeta) shouldRetry(err error) bool {
	msg := strings.ToLower(err.Error())
	if strings.Contains(msg, "too many connections") || strings.Contains(msg, "too many clients") {
		logger.Warnf("transaction failed: %s, will retry it. please increase the max number of connections in your database, or use a connection pool.", msg)
		return true
	}
	return errors.Is(err, errBusy) || strings.Contains(msg, "database is locked")
}

func (m *dbMeta) txn(f func(s *xorm.Session) error, inodes ...Ino) error {
	start := time.Now()

	inodes = []Ino{1}
	var lastErr error
	for i := 0; i < 50; i++ {
		_, err := m.db.Transaction(func(s *xorm.Session) (interface{}, error) {
			return nil, f(s)
		})
		if eno, ok := err.(syscall.Errno); ok && eno == 0 {
			err = nil
		}
		if err != nil && m.shouldRetry(err) {
			logger.Debugf("Transaction failed, restart it (tried %d): %s", i+1, err)
			lastErr = err
			time.Sleep(time.Millisecond * time.Duration(i*i))
			continue
		} else if err == nil && i > 1 {
			logger.Warnf("Transaction succeeded after %d tries (%s), inodes: %v, last error: %s", i+1, time.Since(start), inodes, lastErr)
		}
		return err
	}
	logger.Warnf("Already tried 50 times, returning: %s", lastErr)
	return lastErr
}

func (m *dbMeta) roTxn(f func(s *xorm.Session) error) error {
	start := time.Now()
	s := m.db.NewSession()
	defer s.Close()

	var lastErr error
	for i := 0; i < 50; i++ {
		err := f(s)
		if eno, ok := err.(syscall.Errno); ok && eno == 0 {
			err = nil
		}
		_ = s.Rollback()
		if err != nil && m.shouldRetry(err) {
			logger.Debugf("Read transaction failed, restart it (tried %d): %s", i+1, err)
			lastErr = err
			time.Sleep(time.Millisecond * time.Duration(i*i))
			continue
		} else if err == nil && i > 1 {
			logger.Warnf("Read transaction succeeded after %d tries (%s), last error: %s", i+1, time.Since(start), lastErr)
		}
		return err
	}
	logger.Warnf("Already tried 50 times, returning: %s", lastErr)
	return lastErr
}

func (m *dbMeta) GetNextInode(ctx context.Context, lastIno *Ino) error {
	return m.roTxn(func(s *xorm.Session) error {
		var n node
		if _, err := s.Desc("Inode").Get(&n); err != nil {
			return err
		}
		ino := n.Inode
		ino++
		*lastIno = ino
		return nil
	})
}

func (m *dbMeta) GetAttr(ctx context.Context, inode Ino, attr *Attr) syscall.Errno {
	return errno(m.roTxn(func(s *xorm.Session) error {
		var n = node{Inode: inode}
		ok, err := s.Get(&n)
		if err != nil {
			return err
		} else if !ok {
			return syscall.ENOENT
		}
		m.parseAttr(&n, attr)
		return nil
	}))
}

func (m *dbMeta) SetAttr(ctx context.Context, inode Ino, in *fuse.SetAttrIn, attr *Attr) syscall.Errno {
	return errno(m.txn(func(s *xorm.Session) error {
		var cur = node{Inode: inode}
		ok, err := s.Get(&cur)
		if err != nil {
			return err
		}
		if !ok {
			return syscall.ENOENT
		}
		var curAttr Attr
		m.parseAttr(&cur, &curAttr)
		now := time.Now()

		set := uint16(in.Valid)
		dirtyAttr, st := m.mergeAttr(ctx, inode, set, &curAttr, attr, now)
		if st != 0 {
			return st
		}
		if dirtyAttr == nil {
			return nil
		}

		var dirtyNode node
		m.parseNode(dirtyAttr, &dirtyNode)
		dirtyNode.Ctime = now.UnixNano() / 1e3
		dirtyNode.Ctimensec = int16(now.Nanosecond() % 1000)
		_, err = s.Cols("flags", "mode", "atime", "mtime", "ctime",
			"atimensec", "mtimensec", "ctimensec").
			Update(&dirtyNode, &node{Inode: inode})
		if err == nil {
			m.parseAttr(&dirtyNode, attr)
		}
		return err
	}, inode))
}

func (m *dbMeta) mergeAttr(ctx context.Context, inode Ino, set uint16, cur, attr *Attr, now time.Time) (*Attr, syscall.Errno) {
	// do not allow to change uid, gid or mode. Only meta attributes of time can be changed.
	dirtyAttr := *cur
	var uid uint32
	if fuseCtx, ok := ctx.(*fuse.Context); ok {
		uid = fuseCtx.Uid
	}
	var changed bool
	if set&SetAttrAtimeNow != 0 || (set&SetAttrAtime) != 0 && attr.Atime < 0 {
		/*if st := m.Access(ctx, inode, MODE_MASK_W, cur); ctx.Uid() != cur.Uid && st != 0 {
			return nil, syscall.EACCES
		}*/
		dirtyAttr.Atime = now.Unix()
		dirtyAttr.Atimensec = uint32(now.Nanosecond())
		changed = true
	} else if set&SetAttrAtime != 0 && (cur.Atime != attr.Atime || cur.Atimensec != attr.Atimensec) {
		if uid == 0 {
			return nil, syscall.EPERM
		}
		/*if st := m.Access(ctx, inode, MODE_MASK_W, cur); ctx.Uid() != cur.Uid && st != 0 {
			return nil, syscall.EACCES
		}*/
		dirtyAttr.Atime = attr.Atime
		dirtyAttr.Atimensec = attr.Atimensec
		changed = true
	}
	if set&SetAttrMtimeNow != 0 || (set&SetAttrMtime) != 0 && attr.Mtime < 0 {
		/*if st := m.Access(ctx, inode, MODE_MASK_W, cur); ctx.Uid() != cur.Uid && st != 0 {
			return nil, syscall.EACCES
		}*/
		dirtyAttr.Mtime = now.Unix()
		dirtyAttr.Mtimensec = uint32(now.Nanosecond())
		changed = true
	} else if set&SetAttrMtime != 0 && (cur.Mtime != attr.Mtime || cur.Mtimensec != attr.Mtimensec) {
		if uid == 0 {
			return nil, syscall.EPERM
		}
		/*if st := m.Access(ctx, inode, MODE_MASK_W, cur); ctx.Uid() != cur.Uid && st != 0 {
			return nil, syscall.EACCES
		}*/
		dirtyAttr.Mtime = attr.Mtime
		dirtyAttr.Mtimensec = attr.Mtimensec
		changed = true
	}
	if !changed {
		*attr = *cur
		return nil, 0
	}
	return &dirtyAttr, 0
}

func (m *dbMeta) getNode(parent Ino, name string, nn *namedNode) syscall.Errno {
	return errno(m.roTxn(func(s *xorm.Session) error {
		var edge = edge{Parent: parent, Name: []byte(name)}
		exist, err := s.Get(&edge)
		if err != nil {
			log.Fatalf("Failed to find nodes: %v", err)
			return err
		} else if !exist {
			return syscall.ENOENT
		}
		var node node
		exist, err = s.Where("inode = ?", edge.Inode).Get(&node)
		if err != nil {
			log.Fatalf("Failed to find nodes: %v", err)
			return err
		} else if !exist {
			return syscall.ENOENT
		}
		*nn = namedNode{node: node, Name: []byte(name)}
		return nil
	}))
}

func (m *dbMeta) Lookup(ctx context.Context, parent Ino, name string, inode *Ino, attr *Attr) syscall.Errno {
	nn := namedNode{}
	err := m.getNode(parent, name, &nn)
	if err != 0 {
		return err
	}
	*inode = nn.Inode
	m.parseAttr(&nn.node, attr)
	return 0
}

func (m *dbMeta) Mknod(ctx context.Context, parent Ino, name string, _type uint8, mode uint32, inode *Ino, key []byte, attr *Attr) syscall.Errno {
	return errno(m.txn(func(s *xorm.Session) error {
		var pn = node{Inode: parent}
		ok, err := s.Get(&pn)
		if err != nil {
			return err
		}
		if !ok {
			return syscall.ENOENT
		}
		if pn.Type != TypeDirectory {
			return syscall.ENOTDIR
		}
		var pattr Attr
		m.parseAttr(&pn, &pattr)
		var e = edge{Parent: parent, Name: []byte(name)}
		ok, err = s.Get(&e)
		if err != nil {
			return err
		}
		var foundIno Ino
		var foundType uint8
		if ok {
			foundType, foundIno = e.Type, e.Inode
		}
		if foundIno != 0 {
			if _type == TypeFile || _type == TypeDirectory {
				foundNode := node{Inode: foundIno}
				ok, err = s.Get(&foundNode)
				if err != nil {
					return err
				} else if ok {
					m.parseAttr(&foundNode, attr)
				} else if attr != nil {
					*attr = Attr{Typ: foundType, Parent: parent} // corrupt entry
				}
				*inode = foundIno
			}
			return syscall.EEXIST
		}

		n := node{Inode: *inode}
		if attr != nil {
			m.parseNode(attr, &n) // do almost nothing here (attr is empty)
		}
		mode &= 07777

		var updateParent bool
		now := time.Now().UnixNano()
		if _type == TypeDirectory {
			pn.Nlink++
			updateParent = true
		}
		if updateParent || time.Duration(now-pn.Mtime*1e3-int64(pn.Mtimensec)) >= SkipDirMtime {
			pn.Mtime = now / 1e3
			pn.Ctime = now / 1e3
			updateParent = true
		}
		n.Atime = now / 1e3
		n.Mtime = now / 1e3
		n.Ctime = now / 1e3
		n.Atimensec = int16(now % 1e3)
		n.Mtimensec = int16(now % 1e3)
		n.Ctimensec = int16(now % 1e3)
		n.Parent = parent
		if _type == TypeDirectory {
			n.Nlink = 2
			n.Mode |= 0755
			n.Length = 4 << 10 // 4KB
			n.Type = TypeDirectory
		} else if _type == TypeFile {
			n.Nlink = 1
			n.Length = 0
			n.Mode |= 0644
			n.Rdev = 0
			n.Type = TypeFile
		}

		if err = mustInsert(s, &edge{Parent: parent, Name: []byte(name), Inode: *inode, Type: _type, Key: key}, &n); err != nil {
			return err
		}
		if updateParent {
			if _, err := s.Cols("nlink", "mtime", "ctime", "mtimensec", "ctimensec").Update(&pn, &node{Inode: pn.Inode}); err != nil {
				return err
			}
		}
		m.parseAttr(&n, attr)
		return nil
	}, parent))
}

func (m *dbMeta) joinNodes(parent Ino, nns *[]namedNode) syscall.Errno {
	return errno(m.roTxn(func(s *xorm.Session) error {
		var nodes []node
		err := s.SQL("SELECT * FROM `nsfs_edge` INNER JOIN `nsfs_node` ON nsfs_edge.inode=nsfs_node.inode WHERE nsfs_edge.parent = ?", parent).Find(&nodes)
		if err != nil {
			log.Fatalf("Failed to find nodes: %v", err)
		}
		var edges []edge
		err = s.SQL("SELECT * FROM `nsfs_edge` INNER JOIN `nsfs_node` ON nsfs_edge.inode=nsfs_node.inode WHERE nsfs_edge.parent = ?", parent).Find(&edges)
		if err != nil {
			log.Fatalf("Failed to find edges: %v", err)
		}
		if len(nodes) != len(edges) {
			log.Fatalf("Nodes and edges are not equal: %d %d", len(nodes), len(edges))
		}
		for _, n := range nodes {
			nn := namedNode{node: n}
			for _, e := range edges {
				if e.Inode == n.Inode {
					nn.Name = e.Name
					break
				}
			}
			*nns = append(*nns, nn)
		}
		return nil
	}))
}

func (m *dbMeta) Readdir(ctx context.Context, inode Ino, plus uint8, entries *[]*Entry) syscall.Errno {
	/*s = s.Table(&edge{})
	if plus != 0 {
		s = s.Join("INNER", &node{}, "nsfs_edge.inode=nsfs_node.inode")
	}
	var nodes []namedNode
	if err := s.Find(&nodes, &edge{Parent: inode}); err != nil {
		return err
	}*/
	// The join does not seem to work properly so doing some "brute force"
	nodes := make([]namedNode, 0)
	err := m.joinNodes(inode, &nodes)
	for _, n := range nodes {
		if len(n.Name) == 0 {
			logger.Errorf("Corrupt entry with empty name: inode %d parent %d", n.Inode, inode)
			continue
		}
		entry := &Entry{
			Inode: n.Inode,
			Name:  n.Name,
			Attr:  &Attr{},
		}
		if plus != 0 {
			m.parseAttr(&n.node, entry.Attr)
		} else {
			entry.Attr.Typ = n.Type
		}
		*entries = append(*entries, entry)
	}
	return err
}

func (m *dbMeta) Rmdir(ctx context.Context, parent Ino, name string) syscall.Errno {
	return errno(m.txn(func(s *xorm.Session) error {
		var pn = node{Inode: parent}
		ok, err := s.Get(&pn)
		if err != nil {
			return err
		}
		if !ok {
			return syscall.ENOENT
		}
		if pn.Type != TypeDirectory {
			return syscall.ENOTDIR
		}
		var pattr Attr
		m.parseAttr(&pn, &pattr)
		/*if st := m.Access(ctx, parent, MODE_MASK_W|MODE_MASK_X, &pattr); st != 0 {
			return st
		}*/
		var e = edge{Parent: parent, Name: []byte(name)}
		ok, err = s.Get(&e)
		if err != nil {
			return err
		}
		if !ok {
			return syscall.ENOENT
		}
		if e.Type != TypeDirectory {
			return syscall.ENOTDIR
		}
		var n = node{Inode: e.Inode}
		ok, err = s.Get(&n)
		if err != nil {
			return err
		}
		exist, err := s.Exist(&edge{Parent: e.Inode})
		if err != nil {
			return err
		}
		if exist {
			return syscall.ENOTEMPTY
		}
		now := time.Now().UnixNano()
		if ok {
			/*if ctx.Uid() != 0 && pn.Mode&01000 != 0 && ctx.Uid() != pn.Uid && ctx.Uid() != n.Uid {
				return syscall.EACCES
			}*/
		} else {
			logger.Warnf("no attribute for inode %d (%d, %s)", e.Inode, parent, name)
		}
		pn.Nlink--
		pn.Mtime = now / 1e3
		pn.Ctime = now / 1e3
		pn.Mtimensec = int16(now % 1e3)
		pn.Ctimensec = int16(now % 1e3)

		if _, err := s.Delete(&edge{Parent: parent, Name: e.Name}); err != nil {
			return err
		}

		if _, err := s.Delete(&node{Inode: e.Inode}); err != nil {
			return err
		}

		_, err = s.Cols("nlink", "mtime", "ctime", "mtimensec", "ctimensec").Update(&pn, &node{Inode: pn.Inode})
		return err
	}, parent))
}

func (m *dbMeta) Unlink(ctx context.Context, parent Ino, name string) syscall.Errno {
	return errno(m.txn(func(s *xorm.Session) error {
		var n node
		var pn = node{Inode: parent}
		ok, err := s.Get(&pn)
		if err != nil {
			return err
		}
		if !ok {
			return syscall.ENOENT
		}
		if pn.Type != TypeDirectory {
			return syscall.ENOTDIR
		}
		/*var pattr Attr
		m.parseAttr(&pn, &pattr)
		if st := m.Access(ctx, parent, MODE_MASK_W|MODE_MASK_X, &pattr); st != 0 {
			return st
		}*/
		var e = edge{Parent: parent, Name: []byte(name)}
		ok, err = s.Get(&e)
		if err != nil {
			return err
		}
		if !ok {
			return syscall.ENOENT
		}
		if e.Type == TypeDirectory {
			return syscall.EPERM
		}

		n = node{Inode: e.Inode}
		ok, err = s.Get(&n)
		if err != nil {
			return err
		}
		now := time.Now().UnixNano()
		if ok {
			n.Ctime = now / 1e3
			n.Ctimensec = int16(now % 1e3)
			n.Nlink--
		} else {
			logger.Warnf("no attribute for inode %d (%d, %s)", e.Inode, parent, name)
		}

		var updateParent bool
		if time.Duration(now-pn.Mtime*1e3-int64(pn.Mtimensec)) >= SkipDirMtime {
			pn.Mtime = now / 1e3
			pn.Ctime = now / 1e3
			pn.Mtimensec = int16(now % 1e3)
			pn.Ctimensec = int16(now % 1e3)
			updateParent = true
		}

		if _, err := s.Delete(&edge{Parent: parent, Name: e.Name}); err != nil {
			return err
		}
		if _, err := s.Delete(&node{Inode: e.Inode}); err != nil {
			return err
		}
		if updateParent {
			if _, err = s.Cols("mtime", "ctime", "mtimensec", "ctimensec").Update(&pn, &node{Inode: pn.Inode}); err != nil {
				return err
			}
		}
		if n.Nlink > 0 {
			if _, err := s.Cols("nlink", "ctime", "ctimensec", "parent").Update(&n, &node{Inode: e.Inode}); err != nil {
				return err
			}
		}
		return err
	}, parent))
}

func (m *dbMeta) Write(ctx context.Context, inode uint64, data []byte, off int64) syscall.Errno {
	ino := Ino(inode)
	return errno(m.txn(func(s *xorm.Session) error {
		nodeAttr := node{Inode: ino}
		ok, err := s.Get(&nodeAttr)
		if err != nil {
			return err
		}
		if !ok {
			return syscall.ENOENT
		}
		if nodeAttr.Type != TypeFile {
			return syscall.EPERM
		}
		newleng := uint64(len(data)) + uint64(off)
		nodeAttr.Length = newleng
		now := time.Now()
		nodeAttr.Mtime = now.UnixNano() / 1e3
		nodeAttr.Mtimensec = int16(now.Nanosecond() % 1e3)

		_, err = s.Cols("length", "mtime", "mtimensec").Update(&nodeAttr, &node{Inode: ino})
		return err
	}, ino))
}

func newSQLMeta(driver, addr string) (Meta, error) {
	engine, err := xorm.NewEngine(driver, addr)
	if err != nil {
		return nil, fmt.Errorf("unable to use data source %s: %s", driver, err)
	}

	start := time.Now()
	if err = engine.Ping(); err != nil {
		return nil, fmt.Errorf("ping database: %s", err)
	}
	if time.Since(start) > time.Millisecond*5 {
		logger.Warnf("The latency to database is too high: %s", time.Since(start))
	}
	engine.DB().SetMaxIdleConns(runtime.GOMAXPROCS(-1) * 2)
	engine.DB().SetConnMaxIdleTime(time.Minute * 5)
	engine.SetTableMapper(names.NewPrefixMapper(engine.GetTableMapper(), "nsfs_"))
	m := &dbMeta{
		db:   engine,
		addr: addr,
		root: RootInode,
	}
	return m, nil
}
