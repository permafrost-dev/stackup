package cache

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	carbon "github.com/golang-module/carbon/v2"
	"github.com/stackup-app/stackup/lib/projectinfo"
	"github.com/stackup-app/stackup/lib/utils"
	bolt "go.etcd.io/bbolt"
)

type Cache struct {
	Db         *bolt.DB
	Enabled    bool
	Name       string
	Path       string
	Filename   string
	DefaultTtl int
	ticker     *time.Ticker
}

type HashAlgorithmType byte

const (
	HashAlgorithmSha1 HashAlgorithmType = iota
	HashAlgorithmSha256
	HashAlgorithmSha384
	HashAlgorithmSha512
)

// creates a new Cache instance. `name` is used to determine the boltdb filename, and `storagePath` is
// used as the path for the db file.  If `name` is empty, it defaults to the name of the current binary.
func New(name string, storagePath string) *Cache {
	if !utils.FileExists(storagePath) {
		os.MkdirAll(storagePath, 0744)
	}

	return (&Cache{Name: name, Enabled: false, Path: storagePath, DefaultTtl: 60}).Init()
}

func (c *Cache) AutoPurgeInterval() time.Duration {
	interval, err := time.ParseDuration("60s")
	if err != nil {
		return time.Minute
	}

	return interval
}

func (c *Cache) Cleanup(removeFile bool) {
	if c.Db != nil {
		c.Db.Close()
	}

	if removeFile && utils.FileExists(c.Filename) {
		os.Remove(c.Filename)
	}
}

func (c *Cache) CreateEntry(keyName string, value string, expiresAt *carbon.Carbon, hash string, algorithm string, updatedAt *carbon.Carbon) *CacheEntry {
	if updatedAt == nil {
		temp := carbon.Now()
		updatedAt = &temp
	}

	if expiresAt == nil {
		temp := carbon.Now().AddMinutes(c.DefaultTtl)
		expiresAt = &temp
	}

	return &CacheEntry{
		Key:       keyName,
		Value:     value,
		Algorithm: algorithm,
		Hash:      hash,
		ExpiresAt: expiresAt.ToIso8601String(),
		UpdatedAt: updatedAt.ToIso8601String(),
	}
}

func (c *Cache) GetBaseKey(key string) string {
	result := key
	suffixes := []string{"_expires_at", "_hash", "_updated_at"}
	for _, suffix := range suffixes {
		result = strings.TrimSuffix(result, suffix)
	}

	return result
}

// The `Init` function in the `Cache` struct is used to initialize the cache by setting up the
// necessary configurations and opening the database connection. Here's a breakdown of what it does:
func (c *Cache) Init() *Cache {
	if c.Db != nil {
		return c
	}

	filename := utils.FsSafeName(c.Name) + ".db"

	if c.Name == "" || strings.TrimSuffix(filename, ".db") == "" {
		cwd, _ := os.Getwd()
		filename = projectinfo.New(os.Args[0], cwd).FsSafeName() + ".db"
	}

	c.Filename = filepath.Join(c.Path, filename)
	if c.Name == "" {
		c.Name = strings.TrimSuffix(filename, ".db")
	}

	db, err := bolt.Open(c.Filename, 0644, bolt.DefaultOptions)
	if err != nil {
		c.Enabled = false
		c.Db = nil
		return c
	}

	c.Db = db

	// create a new project bucket if it doesn't exist
	c.Db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(c.Name))
		if err != nil {
			return fmt.Errorf("create bucket: %s", err)
		}
		return nil
	})

	c.Enabled = true
	c.StartAutoPurge()

	return c
}

func (c *Cache) StartAutoPurge() {
	if c.ticker != nil {
		return
	}

	c.ticker = time.NewTicker(c.AutoPurgeInterval())

	go func() {
		for {
			select {
			case <-c.ticker.C:
				c.purgeExpired()
			}
		}
	}()
}

// The `Get` function in the `Cache` struct is used to retrieve the value of a cache entry with a given
// key. It takes a `key` parameter (string) and returns the corresponding value (string).
func (c *Cache) Get(key string) (*CacheEntry, bool) {
	result := CacheEntry{Value: "", ExpiresAt: ""}

	c.Db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(c.Name))
		bytes := b.Get([]byte(key))
		json.Unmarshal(bytes, &result)
		return nil
	})

	if result.IsExpired() {
		return nil, false
	}

	result.DecodeValue()

	return &result, true
}

// The `purgeExpired` function in the `Cache` struct is used to remove any cache entries that have
// expired. It iterates through all the keys in the cache bucket and checks if each key has expired
// using the `IsExpired` function. If a key is expired, it is deleted from the cache bucket. This
// function ensures that expired cache entries are automatically removed from the cache to free up
// space and maintain cache integrity.
func (c *Cache) purgeExpired() {
	primaryKeys := []string{}

	c.Db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(c.Name))

		b.ForEach(func(k, v []byte) error {
			primaryKeys = append(primaryKeys, string(k))
			return nil
		})

		return nil
	})

	if len(primaryKeys) == 0 {
		return
	}

	c.Db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(c.Name))

		for _, key := range primaryKeys {
			if c.IsExpired(key) {
				b.Delete([]byte(key))
			}
		}

		return nil
	})
}

// The `IsExpired` function in the `Cache` struct is used to check if a cache entry with a given key
// has expired.
func (c *Cache) IsExpired(key string) bool {
	item, found := c.Get(c.GetBaseKey(key))
	if !found {
		return true
	}

	return item.IsExpired()
}

// The `HashMatches` function in the `Cache` struct is used to check if the hash value stored in the
// cache for a given key matches a provided hash value.
func (c *Cache) HashMatches(key string, hash string) bool {
	item, found := c.Get(c.GetBaseKey(key))
	if !found {
		return true
	}

	return item.Hash == hash && hash != ""
}

// The `Set` function in the `Cache` struct is used to set a cache entry with a given key and value. It
// takes three parameters: `key` (string), `value` (any), and `ttlMinutes` (int).
func (c *Cache) Set(key string, value *CacheEntry, ttlMinutes int) {
	c.Db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(c.Name))
		value.EncodeValue()

		code, err := json.Marshal(value)
		if err == nil {
			err = b.Put([]byte(key), code)
		}

		value.DecodeValue()

		return err
	})
}

// The `Has` function in the `Cache` struct is used to check if a cache entry with a given key exists
// and is not expired. It calls the `Get` function to retrieve the value of the cache entry with the
// given key and checks if the value is not empty (`c.Get(key) != ""`) and if the cache entry is not
// expired (`!c.IsExpired(key)`). If both conditions are true, it returns `true`, indicating that the
// cache entry exists and is valid. Otherwise, it returns `false`.
func (c *Cache) Has(key string) bool {
	found := false

	c.Db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(c.Name))
		b.ForEach(func(k, v []byte) error {
			if string(k) == key {
				found = true
			}
			return nil
		})
		return nil
	})

	return found
}

func (c *Cache) HasUnexpired(key string) bool {
	return c.Has(key) && !c.IsExpired(key)
}

// The `Remove` function in the `Cache` struct is used to remove a cache entry with a given key. It
// calls the `Set` function with a `value` parameter of `nil` and a `ttlMinutes` parameter of `0`. This
// effectively sets the cache entry to be empty and expired, effectively removing it from the cache.
func (c *Cache) Remove(key string) {
	if len(key) == 0 {
		return
	}

	c.Db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(c.Name))
		b.Delete([]byte(key))
		return nil
	})
}

// builds a cache key using a prefix and name
func (c *Cache) MakeCacheKey(prefix string, name string) string {
	prefix = strings.TrimSuffix(prefix, ":")
	if len(prefix) == 0 {
		return name
	}

	return prefix + ":" + name
}
