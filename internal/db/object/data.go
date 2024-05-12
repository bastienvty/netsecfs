package object

import (
	"fmt"
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
