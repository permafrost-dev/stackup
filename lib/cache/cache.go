package cache

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	carbon "github.com/golang-module/carbon/v2"
	"github.com/stackup-app/stackup/lib/projectinfo"
	bolt "go.etcd.io/bbolt"
)

type Cache struct {
	Db       *bolt.DB
	Enabled  bool
	Name     string
	Path     string
	Filename string
}

type HashAlgorithmType byte

const (
	HashAlgorithmSha1 HashAlgorithmType = iota
	HashAlgorithmSha256
	HashAlgorithmSha384
	HashAlgorithmSha512
)

type CacheEntry struct {
	Key       string
	Value     string
	ExpiresAt string
	Hash      string
	Algorithm string
	UpdatedAt string
}

func CreateCacheEntry(keyName string, value string, expiresAt *carbon.Carbon, hash string, algorithm string, updatedAt *carbon.Carbon) *CacheEntry {
	if updatedAt == nil {
		temp := carbon.Now()
		updatedAt = &temp
	}

	return &CacheEntry{
		Key:       keyName,
		Value:     value,
		Algorithm: algorithm,
		ExpiresAt: expiresAt.ToDateTimeString("America/New_York"),
		Hash:      hash,
		UpdatedAt: updatedAt.ToDateTimeString("America/New_York"),
	}
}

// The function creates and initializes a cache object.
func CreateCache(name string) *Cache {
	result := Cache{Name: name, Enabled: false}
	result.Init()
	result.Enabled = true

	return &result
}

// The function ensures that a directory with a given name exists in the user's home directory and
// returns the path to the directory.
func ensureConfigDirExists(dirName string) (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	// Append the directory name to the home directory
	configDir := filepath.Join(homeDir, dirName)

	// Ensure the directory exists
	err = os.MkdirAll(configDir, 0744)
	if err != nil {
		return "", err
	}

	return configDir, nil
}

func (ce *CacheEntry) IsExpired() bool {
	return carbon.Parse(ce.ExpiresAt, "America/New_York").IsPast()
}

func (c *Cache) IsBaseKey(key string) bool {
	return !strings.HasSuffix(key, "_expires_at") &&
		!strings.HasSuffix(key, "_hash") &&
		!strings.HasSuffix(key, "_updated_at")
}

func (c *Cache) GetBaseKey(key string) string {
	if c.IsBaseKey(key) {
		return key
	}
	return key
}

// The `Init` function in the `Cache` struct is used to initialize the cache by setting up the
// necessary configurations and opening the database connection. Here's a breakdown of what it does:
func (c *Cache) Init() {
	if c.Db != nil {
		return
	}

	c.Path, _ = ensureConfigDirExists(".stackup")
	c.Filename = filepath.Join(c.Path, "stackup.db")
	if c.Name == "" {
		c.Name = projectinfo.New().FsSafeName()
	}

	db, err := bolt.Open(c.Filename, 0644, bolt.DefaultOptions)
	if err != nil {
		c.Enabled = false
		c.Db = nil
		return
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
}

func (c *Cache) StartAutoPurge() {
	interval, err := time.ParseDuration("15s")
	if err != nil {
		interval = time.Minute
	}

	ticker := time.NewTicker(interval)
	go func() {
		for {
			select {
			case <-ticker.C:
				c.purgeExpired()
			}
		}
	}()
}

// The `Get` function in the `Cache` struct is used to retrieve the value of a cache entry with a given
// key. It takes a `key` parameter (string) and returns the corresponding value (string).
func (c *Cache) Get(key string) (*CacheEntry, bool) {
	result := CacheEntry{}

	if !c.Has(key) {
		return nil, false
	}

	// if c.IsExpired(key) {
	// 	c.Remove(key)
	// 	return nil, false
	// }

	c.Db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(c.Name))
		bytes := b.Get([]byte(key))
		err := json.Unmarshal(bytes, &result)
		if err != nil {
			fmt.Println("error:", err)
		}
		return nil
	})

	tempValue, _ := base64.StdEncoding.DecodeString(result.Value)
	result.Value = string(tempValue)

	if result.IsExpired() {
		return nil, false
	}

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

	c.Db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(c.Name))

		for _, key := range primaryKeys {
			if c.IsExpired(key) {
				fmt.Printf("Removing expired cache entry: %s\n", key)
				b.Delete([]byte(key))
			}
		}

		return nil
	})

	c.Db.Sync()
}

// The `GetExpiresAt` function in the `Cache` struct is used to retrieve the expiration time of a cache
// entry with a given key.
func (c *Cache) GetExpiresAt(key string) *carbon.Carbon {
	value, found := c.Get(c.GetBaseKey(key))

	if !found {
		return nil
	}

	result := carbon.Parse(value.ExpiresAt, "America/New_York")
	return &result
}

func (c *Cache) GetUpdatedAt(key string) *carbon.Carbon {
	value, found := c.Get(c.GetBaseKey(key))

	if !found {
		return nil
	}

	result := carbon.Parse(value.UpdatedAt, "America/New_York")
	return &result
}

// The `IsExpired` function in the `Cache` struct is used to check if a cache entry with a given key
// has expired.
func (c *Cache) IsExpired(key string) bool {
	item, found := c.Get(c.GetBaseKey(key))
	if !found {
		return true
	}

	result := carbon.Parse(item.ExpiresAt, "America/New_York")
	return result.IsPast()
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
		tempValue := *value
		tempValue.Value = base64.StdEncoding.EncodeToString([]byte(value.Value))

		code, err := json.Marshal(tempValue)

		if err != nil {
			fmt.Printf("Error: %s\n", err)
		}

		err = b.Put([]byte(key), code)
		return err
	})

	c.Db.Sync()
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

	c.Db.Sync()
}

// builds a cache key using a prefix and name
func (c *Cache) MakeCacheKey(prefix string, name string) string {
	prefix = strings.TrimSuffix(prefix, ":")
	if len(prefix) == 0 {
		return name
	}

	return prefix + ":" + name
}
