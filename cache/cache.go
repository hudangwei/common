package cache

import "github.com/syndtr/goleveldb/leveldb"

// Cache ...
type Cache struct {
	leveldb *leveldb.DB
}

// NewCache ...
func NewCache(path string) (*Cache, error) {
	db, err := leveldb.OpenFile(path, nil)
	if err != nil {
		return nil, err
	}
	return &Cache{
		leveldb: db,
	}, nil
}

// Has ...
func (c *Cache) Has(key string) bool {
	exist, err := c.leveldb.Has([]byte(key), nil)
	if err != nil {
		return false
	}
	return exist
}

// Get ...
func (c *Cache) Get(key string) ([]byte, error) {
	return c.leveldb.Get([]byte(key), nil)
}

// Set ...
func (c *Cache) Set(key string, v []byte) error {
	return c.leveldb.Put([]byte(key), v, nil)
}

// Delete ...
func (c *Cache) Delete(key string) error {
	return c.leveldb.Delete([]byte(key), nil)
}

// ListKey ...
func (c *Cache) ListKey() ([]string, error) {
	var keys []string
	iter := c.leveldb.NewIterator(nil, nil)
	for iter.Next() {
		keys = append(keys, string(iter.Key()))
	}
	iter.Release()
	if err := iter.Error(); err != nil {
		return nil, err
	}
	return keys, nil
}

// Close ...
func (c *Cache) Close() error {
	return c.leveldb.Close()
}
