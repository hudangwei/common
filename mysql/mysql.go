package mysql

import (
	"errors"
	"fmt"
	"net"
	"sync"
	"sync/atomic"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/hudangwei/common/depends"
	"github.com/hudangwei/common/logger"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

const (
	DefaultConnectWaiTimeout = 15 * time.Second
	DefaultCharset           = "utf8"
	CharsetUtf8mb4           = "utf8mb4"
)

var nilConfigErr = errors.New("config is nil")

type MySqlConfig struct {
	Host         string `toml:"host"`
	Port         int    `toml:"port"`
	User         string `toml:"user"`
	Password     string `toml:"password"`
	DBName       string `toml:"db_name"`
	MaxConns     int    `toml:"max_conns"`
	MaxIdleConns int    `toml:"max_idle_conns"`
}

type Mysql struct {
	db        *sqlx.DB
	closeOnce sync.Once
	closeFlag int32
	closeChan chan struct{}
}

func (m *Mysql) Open(f depends.Configger, name string) error {
	conf, err := f.LoadConfig(&MySqlConfig{}, name)
	if err != nil || conf == nil {
		logger.Error("mysql config with error", zap.Error(err), zap.String("mysql config name", name))
		return err
	}
	err = m.OpenWithConfig(conf.(*MySqlConfig))
	if err != nil {
		logger.Error("mysql open with error", zap.Error(err), zap.String("mysql config name", name), zap.String("mysql addr", conf.(*MySqlConfig).Host))
		return err
	}
	logger.Info("mysql open ok", zap.String("mysql config name", name), zap.String("mysql addr", conf.(*MySqlConfig).Host))

	return nil
}

func (m *Mysql) OpenWithConfig(conf *MySqlConfig) error {
	if conf == nil {
		return nilConfigErr
	}
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?timeout=%s&charset=%s",
		conf.User, conf.Password, conf.Host, conf.Port, conf.DBName, DefaultConnectWaiTimeout.String(), DefaultCharset)
	sqlxDB, err := sqlx.Connect("mysql", dsn)
	if err != nil {
		return err
	}
	m.db = sqlxDB
	if conf.MaxIdleConns > 0 {
		sqlxDB.SetMaxIdleConns(conf.MaxIdleConns)
	}
	if conf.MaxConns > 0 {
		sqlxDB.SetMaxOpenConns(conf.MaxConns)
	}
	m.closeChan = make(chan struct{})
	go m.ping()
	return nil
}

func (m *Mysql) DB() *sqlx.DB {
	return m.db
}

func (m *Mysql) Close() {
	m.closeOnce.Do(func() {
		atomic.StoreInt32(&m.closeFlag, 1)
		close(m.closeChan)
		if m.db != nil {
			m.db.Close()
		}
	})
}

func (m *Mysql) IsClosed() bool {
	return atomic.LoadInt32(&m.closeFlag) == 1
}

func (m *Mysql) ping() {
	for {
		select {
		case <-m.closeChan:
			return
		default:
		}
		if m.IsClosed() {
			return
		}
		db := m.DB()
		err := db.Ping()
		if err != nil {
			if nerr, ok := err.(*net.OpError); ok && nerr.Op == "dial" {
				logger.Error("alarm mysql dial error", zap.Error(err))
			}
		}
		time.Sleep(10 * time.Second)
	}
}
