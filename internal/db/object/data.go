package object

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/sirupsen/logrus"
	"xorm.io/xorm"
	"xorm.io/xorm/log"
	"xorm.io/xorm/names"
)

type dbData struct {
	db   *xorm.Engine
	addr string
}

func (s *dbData) String() string {
	driver := s.db.DriverName()
	return fmt.Sprintf("%s://%s/", driver, s.addr)
}

type blob struct {
	Inode    uint64    `xorm:"pk"`
	Key      []byte    `xorm:"notnull"`
	Size     int64     `xorm:"notnull"`
	Modified time.Time `xorm:"notnull updated"`
	Data     []byte    `xorm:"mediumblob"`
}

func (s *dbData) Get(inode uint64, off int64, key *[]byte) ([]byte, error) {
	var b = blob{Inode: inode}
	ok, err := s.db.Get(&b)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, os.ErrNotExist
	}
	if off > int64(len(b.Data)) {
		off = int64(len(b.Data))
	}
	data := b.Data[off:]
	*key = b.Key
	return data, nil
}

func (s *dbData) Put(inode uint64, key []byte, data []byte, size int64) error {
	now := time.Now()
	// size of clear data (not encrypted) -> TODO: update length of encrypted data
	b := blob{Inode: inode, Key: key, Data: data, Size: size, Modified: now}
	n, err := s.db.Insert(&b)
	if err != nil || n == 0 {
		n, err = s.db.Update(&b, &blob{Inode: inode})
	}
	if err == nil && n == 0 {
		err = errors.New("not inserted or updated")
	}
	return err
}

func (s *dbData) Delete(inode uint64, key string) error {
	affected, err := s.db.Delete(&blob{Inode: inode})
	if err == nil && affected == 0 {
		return nil
	}
	return err
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
