package db

import (
	"context"
	"fmt"
	"runtime"
	"syscall"
	"time"

	"github.com/bastienvty/netsecfs/utils"
	"github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"
	_ "github.com/mattn/go-sqlite3"
	"xorm.io/xorm"
	"xorm.io/xorm/names"
)

var logger = utils.GetLogger("juicefs")

type Ino uint64

type node struct {
	Inode     Ino    `xorm:"pk"`
	Type      uint8  `xorm:"notnull"`
	Flags     uint8  `xorm:"notnull"`
	Mode      uint16 `xorm:"notnull"`
	Uid       uint32 `xorm:"notnull"`
	Gid       uint32 `xorm:"notnull"`
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

type user struct {
	Id       uint32 `xorm:"pk autoincr"`
	Username string `xorm:"notnull unique"`
	Password string `xorm:"notnull"`
}

type Meta interface {
	Name() string
	Init() error
	Shutdown() error
	GetAttr(ctx context.Context, f fs.FileHandle, out *fuse.AttrOut) (ch *fs.Inode, errno syscall.Errno)
}

type dbMeta struct {
	db   *xorm.Engine
	addr string
}

func (m *dbMeta) Name() string {
	return m.addr
}

func (m *dbMeta) Init() error {
	if err := m.db.Sync2(new(node), new(user)); err != nil {
		return fmt.Errorf("sync tables: %s", err)
	}
	return nil
}

func (m *dbMeta) Shutdown() error {
	return m.db.Close()
}

/*func (m *dbMeta) parseAttr(buf []byte, attr *fuse.Attr) {
	if attr == nil || len(buf) == 0 {
		return
	}
	rb := utils.FromBuffer(buf)
	attr.Flags = rb.Get8()
	attr.Mode = rb.Get16()
	attr.Typ = uint8(attr.Mode >> 12)
	attr.Mode &= 0xfff
	attr.Uid = rb.Get32()
	attr.Gid = rb.Get32()
	attr.Atime = int64(rb.Get64())
	attr.Atimensec = rb.Get32()
	attr.Mtime = int64(rb.Get64())
	attr.Mtimensec = rb.Get32()
	attr.Ctime = int64(rb.Get64())
	attr.Ctimensec = rb.Get32()
	attr.Nlink = rb.Get32()
	attr.Length = rb.Get64()
	attr.Rdev = rb.Get32()
	if rb.Left() >= 8 {
		attr.Parent = Ino(rb.Get64())
	}
	attr.Full = true
	if rb.Left() >= 8 {
		attr.AccessACL = rb.Get32()
		attr.DefaultACL = rb.Get32()
	}
	logger.Tracef("attr: %+v -> %+v", buf, attr)
}*/

func (m *dbMeta) GetAttr(ctx context.Context, f fs.FileHandle, out *fuse.AttrOut) (ch *fs.Inode, errno syscall.Errno) {
	return nil, syscall.ENOSYS
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
	}
	return m, nil
}

func RegisterMeta(addr string) Meta {
	m, err := newSQLMeta("sqlite3", addr)
	if err != nil {
		logger.Fatalf("unable to register client: %s", err)
	}
	return m
}
