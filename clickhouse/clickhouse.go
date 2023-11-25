package clickhouse

import (
	"errors"
	"fmt"

	_ "github.com/ClickHouse/clickhouse-go"
	"github.com/jmoiron/sqlx"
)

var nilConfigErr = errors.New("config is nil")

type ClickHouseConfig struct {
	Host     string `toml:"host"`
	Port     int    `toml:"port"`
	User     string `toml:"user"`
	Password string `toml:"password"`
	Debug    bool   `toml:"debug"`
	DbName   string `toml:"dbname"`
}

type ClickHouse struct {
	db *sqlx.DB
}

func (m *ClickHouse) OpenWithConfig(conf *ClickHouseConfig) error {
	if conf == nil {
		return nilConfigErr
	}
	dsn := fmt.Sprintf("tcp://%s:%d?debug=%v&username=%s&password=%s&database=%s",
		conf.Host, conf.Port, conf.Debug, conf.User, conf.Password, conf.DbName)
	db, err := sqlx.Open("clickhouse", dsn)
	if err != nil {
		return err
	}
	err = db.Ping()
	if err != nil {
		return err
	}
	m.db = db
	return nil
}

func (m *ClickHouse) DB() *sqlx.DB {
	return m.db
}
