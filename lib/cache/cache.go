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
func New(name string, storagePath string, ttlMinutes int) *Cache {
	if !utils.FileExists(storagePath) {
		os.MkdirAll(storagePath, 0744)
	}

	result := Cache{Name: name, Enabled: false, Path: storagePath, DefaultTtl: ttlMinutes}

	return result.Initialize()
}

func NewCacheEntry(obj any, ttlMinutes int) *CacheEntry {
	updatedObj := carbon.Now()
	expiresObj := carbon.Now().AddMinutes(ttlMinutes)

	value, err := json.Marshal(obj)

	if err != nil {
		return nil
	}

	return &CacheEntry{
		Value:     string(value),
		Algorithm: "",
		Hash:      "",
		ExpiresAt: expiresObj.ToIso8601String(),
		UpdatedAt: updatedObj.ToIso8601String(),
	}
}

func CreateCarbonNowPtr() *carbon.Carbon {
	result := carbon.Now()
	return &result
}

func CreateExpiresAtPtr(ttlMinutes int) *carbon.Carbon {
	result := CreateCarbonNowPtr().AddMinutes(ttlMinutes)

	return &result
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
		c.Db = nil
	}

	if removeFile && utils.FileExists(c.Filename) {
		os.Remove(c.Filename)
	}
}

func (c *Cache) CreateEntry(value string, expiresAt *carbon.Carbon, hash string, algorithm string, updatedAt *carbon.Carbon) *CacheEntry {
	updatedObj := carbon.Now()
	expiresObj := carbon.Now().AddMinutes(c.DefaultTtl)

	if updatedAt != nil {
		updatedObj = *updatedAt
	}

	if expiresAt != nil {
		expiresObj = *expiresAt
	}

	return &CacheEntry{
		Value:     value,
		Algorithm: algorithm,
		Hash:      hash,
		ExpiresAt: expiresObj.ToIso8601String(),
		UpdatedAt: updatedObj.ToIso8601String(),
	}
}

func (c *Cache) MakeKey(key string) string {
	result := key
	suffixes := []string{"_expires_at", "_hash", "_updated_at"}
	for _, suffix := range suffixes {
		result = strings.TrimSuffix(result, suffix)
	}

	return result
}

// The `Initialize` function in the `Cache` struct is used to initialize the cache by setting up the
// necessary configurations and opening the database connection. Here's a breakdown of what it does:
func (c *Cache) Initialize() *Cache {
	if c.Db != nil {
		return c
	}

	filename := utils.EnforceSuffix(utils.FsSafeName(c.Name), ".db")

	if strings.TrimSuffix(filename, ".db") == "" {
		cwd, _ := os.Getwd()
		filename = utils.EnforceSuffix(projectinfo.New(os.Args[0], cwd).FsSafeName(), ".db")
	}

	if c.Name == "" {
		c.Name = strings.TrimSuffix(filename, ".db")
	}

	c.Filename = filepath.Join(c.Path, filename)

	var err error
	if c.Db, err = bolt.Open(c.Filename, 0644, &bolt.Options{Timeout: 5 * time.Second}); err != nil {
		return c
	}

	// create a new project bucket if it does not already exist
	c.Db.Update(func(tx *bolt.Tx) error {
		if _, err := tx.CreateBucketIfNotExists([]byte(c.Name)); err != nil {
			return fmt.Errorf("error creating cache bucket: %s", err)
		}
		return nil
	})

	c.Enabled = true
	c.startAutoPurge()

	return c
}

func (c *Cache) startAutoPurge() {
	// prevent multiple tickers from being created
	if c.ticker != nil {
		return
	}

	// perform an initial purge
	c.purgeExpired()
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
// returns a valid, unexpired cache item or nil if the item is expired or not found.
// note: do not call `Cache.Has()` from here.
func (c *Cache) Get(key string) (*CacheEntry, bool) {
	var err error
	entry := &CacheEntry{}

	c.Db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(c.Name))
		bytes := bucket.Get([]byte(c.MakeKey(key)))
		err = json.Unmarshal(bytes, &entry)

		if entry != nil {
			entry.UpdateTimestampsFromStrings()
			entry.DecodeValue()
		}

		return nil
	})

	// return nothing if there was an error or the entry was found, but is expired
	if err != nil || entry.IsExpired() {
		return nil, false
	}

	// return a valid, unexpired cache item
	return entry, true
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
	item, found := c.Get(key)
	if !found {
		return true
	}

	return item.IsExpired()
}

// The `Set` function in the `Cache` struct is used to set a cache entry with a given key and value. It
// takes three parameters: `key` (string), `value` (any), and `ttlMinutes` (int).
// the .Value attribute of `value.Value` attribute is automatically base64 encoded before being stored.
func (c *Cache) Set(key string, value *CacheEntry, ttlMinutes int) {
	c.Db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(c.Name))

		value.EncodeValue()
		defer value.DecodeValue()

		var err error
		if code, err := json.Marshal(value); err == nil {
			err = b.Put([]byte(key), code)
		}

		return err
	})
}

// The `Has` function in the `Cache` struct is used to check if a cache entry with a given key exists
// and is not expired.
func (c *Cache) Has(key string) bool {
	found := false

	// c.Db.View(func(tx *bolt.Tx) error {
	// b := tx.Bucket([]byte(c.Name))
	//v := b.Get([]byte(key))
	item, ok := c.Get(key)
	if ok {
		found = !item.IsExpired()
	}
	// return nil
	// })

	return found
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
