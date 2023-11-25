package redis

import (
	"errors"
	"net"
	"strconv"
	"time"

	redigo "github.com/gomodule/redigo/redis"
	"github.com/hudangwei/common/logger"
	"go.uber.org/zap"
)

var (
	_nilConfigErr = errors.New("config is nil")
)

type RedisConfig struct {
	Addr  string `toml:"addr"`
	Index int    `toml:"index"`
	Pwd   string `toml:"pwd"`
	Size  int    `toml:"size"`
}

type Redis struct {
	pool *redigo.Pool
}

func (r *Redis) OpenWithConfig(conf *RedisConfig) error {
	r.pool = NewPool(conf)
	return nil
}

func NewPool(conf *RedisConfig) *redigo.Pool {
	var opts []redigo.DialOption
	opts = append(opts, redigo.DialConnectTimeout(2*time.Second))
	opts = append(opts, redigo.DialWriteTimeout(2*time.Second))
	opts = append(opts, redigo.DialReadTimeout(2*time.Second))
	if conf.Index != 0 {
		opts = append(opts, redigo.DialDatabase(conf.Index))
	}
	if len(conf.Pwd) > 0 {
		opts = append(opts, redigo.DialPassword(conf.Pwd))
	}
	dial := func() (redigo.Conn, error) {
		return redigo.Dial("tcp", conf.Addr, opts...)
	}
	if conf.Size == 0 {
		conf.Size = 1
	}
	pool := redigo.NewPool(dial, conf.Size)
	pool.MaxActive = 20
	pool.IdleTimeout = time.Second
	pool.Wait = true
	testOnBorrow := func(c redigo.Conn, t time.Time) error {
		if time.Since(t) < 30*time.Second {
			return nil
		}
		_, err := c.Do("PING")
		if err != nil {
			if nerr, ok := err.(*net.OpError); ok && nerr.Op == "dial" {
				logger.Error("alarm redis dial error", zap.Error(err))
			}
		}
		return err
	}
	pool.TestOnBorrow = testOnBorrow

	return pool
}

func (r *Redis) Do(command string, args ...interface{}) (interface{}, error) {
	c := r.pool.Get()
	defer c.Close()
	resp, err := c.Do(command, args...)
	if err != nil {
		if nerr, ok := err.(*net.OpError); ok && nerr.Op == "dial" {
			logger.Error("alarm redis dial error", zap.Error(err))
		}
	}
	return resp, err
}

// --------
func (r *Redis) Get(key string) (interface{}, error) {
	return r.Do("GET", key)
}

func (r *Redis) Set(key string, value interface{}) error {
	_, err := r.Do("SET", key, value)
	return err
}

func (r *Redis) SetEx(key string, value interface{}, timeout int64) error {
	_, err := r.Do("SETEX", key, timeout, value)
	return err
}

func (r *Redis) Del(keys ...string) (int64, error) {
	ks := make([]interface{}, len(keys))
	for i, key := range keys {
		ks[i] = key
	}
	return redigo.Int64(r.Do("DEL", ks...))
}

func (r *Redis) Incr(key string) (int64, error) {
	return redigo.Int64(r.Do("INCR", key))
}

func (r *Redis) Expire(key string, duration int64) (int64, error) {
	return redigo.Int64(r.Do("EXPIRE", key, duration))
}

func (r *Redis) TTl(key string) (int64, error) {
	return redigo.Int64(r.Do("TTL", key))
}

// ------------
func (r *Redis) HSet(key string, field string, val interface{}) error {
	_, err := r.Do("HSET", key, field, val)
	return err
}

func (r *Redis) HGet(key string, field string) (interface{}, error) {
	return r.Do("HGET", key, field)
}

func (r *Redis) HIncrBy(key, field string, delta int64) (int64, error) {
	return redigo.Int64(r.Do("HINCRBY", key, field, delta))
}

func (r *Redis) HDel(key string, fields ...string) (int64, error) {
	ks := make([]interface{}, len(fields)+1)
	ks[0] = key
	for i, key := range fields {
		ks[i+1] = key
	}
	return redigo.Int64(r.Do("HDEL", ks...))
}

// ------------
func (r *Redis) SAdd(key string, members ...interface{}) (int64, error) {
	args := make([]interface{}, len(members)+1)
	args[0] = key
	for i, m := range members {
		args[i+1] = m
	}
	return redigo.Int64(r.Do("SADD", args...))
}

func (r *Redis) SRem(key string, members ...interface{}) (int64, error) {
	args := make([]interface{}, len(members)+1)
	args[0] = key
	for i, m := range members {
		args[1+i] = m
	}
	return redigo.Int64(r.Do("SREM", args...))
}

func (r *Redis) SIsMember(key string, member interface{}) (bool, error) {
	return redigo.Bool(r.Do("SISMEMBER", key, member))
}

func (r *Redis) SMembers(key string) ([]string, error) {
	values, err := redigo.Strings(r.Do("SMEMBERS", key))
	if err != nil {
		return nil, err
	}
	return values, nil
}

// ------------
func (r *Redis) ZRangeByScoreWithScoreLimited(key string, min, max int64, offset, count int64) ([]string, map[string]int64, error) {
	vals, err := redigo.Values(r.Do("ZRANGEBYSCORE", key, min, max, "WITHSCORES", "LIMIT", offset, count))
	if err != nil {
		return nil, nil, err
	}
	n := len(vals) / 2
	list := make([]string, n)
	res := make(map[string]int64, n)
	for i := 0; i < n; i++ {
		key, _ := redigo.String(vals[2*i], nil)
		score, _ := redigo.String(vals[2*i+1], nil)
		v, _ := strconv.ParseFloat(score, 64)
		res[key] = int64(v)
		list[i] = key
	}
	return list, res, nil
}

func (r *Redis) ZAdd(key string, kvs ...interface{}) (int64, error) {
	if len(kvs) == 0 {
		return 0, nil
	}
	if len(kvs)%2 != 0 {
		return 0, errors.New("args num error")
	}
	args := make([]interface{}, len(kvs)+1)
	args[0] = key
	for i := 0; i < len(kvs); i += 2 {
		args[i+1] = kvs[i]
		args[i+2] = kvs[i+1]
	}
	return redigo.Int64(r.Do("ZAdd", args...))
}

// ------------
func (r *Redis) LPush(key string, value ...interface{}) error {
	args := make([]interface{}, len(value)+1)
	args[0] = key
	for i, v := range value {
		args[i+1] = v
	}
	_, err := r.Do("LPUSH", args...)
	return err
}
