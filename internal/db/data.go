package db

import (
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
	"xorm.io/xorm"
	"xorm.io/xorm/log"
	"xorm.io/xorm/names"
)

type Object interface {
	Key() string
	Size() int64
	Mtime() time.Time
	IsDir() bool
	IsSymlink() bool
	StorageClass() string
}

type obj struct {
	key   string
	size  int64
	mtime time.Time
	isDir bool
	sc    string
}

func (o *obj) Key() string          { return o.key }
func (o *obj) Size() int64          { return o.size }
func (o *obj) Mtime() time.Time     { return o.mtime }
func (o *obj) IsDir() bool          { return o.isDir }
func (o *obj) IsSymlink() bool      { return false }
func (o *obj) StorageClass() string { return o.sc }

// ObjectStorage is the interface for object storage.
// all of these API should be idempotent.
type ObjectStorage interface {
	// Description of the object storage.
	String() string
	// Get the data for the given object specified by key.
	/*Get(key string, off, limit int64, getters ...AttrGetter) (io.ReadCloser, error)
	// Put data read from a reader to an object specified by key.
	Put(key string, in io.Reader, getters ...AttrGetter) error
	// Copy an object from src to dst.
	Copy(dst, src string) error
	// Delete a object.
	Delete(key string, getters ...AttrGetter) error*/
}

type dbData struct {
	db   *xorm.Engine
	addr string
}

func (d *dbData) String() string {
	return d.addr
}

type blob struct {
	Id       int64     `xorm:"pk bigserial"`
	Key      []byte    `xorm:"notnull unique(blob) varbinary(255) "`
	Size     int64     `xorm:"notnull"`
	Modified time.Time `xorm:"notnull updated"`
	Data     []byte    `xorm:"mediumblob"`
}

func newSQLStore(driver, addr string) (ObjectStorage, error) {
	engine, err := xorm.NewEngine(driver, addr)
	if err != nil {
		return nil, fmt.Errorf("open %s: %s", addr, err)
	}
	switch logger.Level { // make xorm less verbose
	case logrus.TraceLevel:
		engine.SetLogLevel(log.LOG_DEBUG)
	case logrus.DebugLevel:
		engine.SetLogLevel(log.LOG_INFO)
	case logrus.InfoLevel, logrus.WarnLevel:
		engine.SetLogLevel(log.LOG_WARNING)
	case logrus.ErrorLevel:
		engine.SetLogLevel(log.LOG_ERR)
	default:
		engine.SetLogLevel(log.LOG_OFF)
	}
	engine.SetTableMapper(names.NewPrefixMapper(engine.GetTableMapper(), "nsfs_"))
	if err := engine.Sync2(new(blob)); err != nil {
		return nil, fmt.Errorf("create table blob: %s", err)
	}
	return &dbData{engine, addr}, nil
}

func CreateStorage(addr string) (ObjectStorage, error) {
	return newSQLStore("sqlite3", addr)
}
