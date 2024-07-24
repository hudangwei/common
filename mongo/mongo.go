package mongo

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"time"

	"github.com/hudangwei/common/depends"
	"github.com/hudangwei/common/logger"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"go.uber.org/zap"
)

var errConfigIsNil = errors.New("config is nil")

type MongoConfig struct {
	Host        string `toml:"host"`
	Port        int    `toml:"port"`
	User        string `toml:"user"`
	Password    string `toml:"password"`
	DBName      string `toml:"db_name"`
	AuthSource  string `toml:"auth_source"`
	MaxConns    int    `toml:"max_conns"`
	MaxIdleTime int    `toml:"max_idle_time"`
}

type Mongo struct {
	db *mongo.Client
}

func (m *Mongo) Open(f depends.Configger, name string) error {
	conf, err := f.LoadConfig(&MongoConfig{}, name)
	if err != nil || conf == nil {
		logger.Error("mongo config with error", zap.Error(err), zap.String("mongo config name", name))
		return err
	}
	err = m.OpenWithConfig(conf.(*MongoConfig))
	if err != nil {
		logger.Error("mongo open with error", zap.Error(err), zap.String("mongo config name", name), zap.String("mongo addr", conf.(*MongoConfig).Host))
		return err
	}
	logger.Info("mongo open ok", zap.String("mongo config name", name), zap.String("mongo addr", conf.(*MongoConfig).Host))

	return nil
}

func (m *Mongo) OpenWithConfig(conf *MongoConfig) error {
	if conf == nil {
		return errConfigIsNil
	}
	dsn := buildDsn(conf.User, conf.Password, conf.Host, conf.Port, conf.DBName, conf.AuthSource)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	clientOptions := options.Client()
	clientOptions.ApplyURI(dsn)
	clientOptions.SetMaxConnIdleTime(time.Duration(conf.MaxIdleTime) * time.Second)
	if conf.MaxConns == 0 {
		conf.MaxConns = 100
	}
	clientOptions.SetMaxPoolSize(uint64(conf.MaxConns))

	newdb, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return err
	}
	err = newdb.Ping(ctx, readpref.Primary())
	if err != nil {
		return err
	}
	m.db = newdb
	return nil
}

// mongodb://username:password@host:port
func buildDsn(username, password, host string, port int, dbName, authSource string) (dsn string) {
	dsn = "mongodb://"

	if username != "" {
		dsn += username
		if password != "" {
			dsn += ":" + password
		}
		dsn += "@"
	}

	dsn += host + ":" + fmt.Sprint(port) + "/" + dbName

	if authSource != "" {
		values := url.Values{}
		values.Add("authSource", authSource)
		dsn = fmt.Sprintf("%s?%s", dsn, values.Encode())
	}
	return dsn
}

func (m *Mongo) DB() *mongo.Client {
	return m.db
}
