package cache

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	carbon "github.com/golang-module/carbon/v2"
	"github.com/stackup-app/stackup/lib/utils"
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
	return &CacheEntry{
		Key:       keyName,
		Value:     value,
		Algorithm: algorithm,
		ExpiresAt: expiresAt.ToDateTimeString(),
		Hash:      hash,
		UpdatedAt: updatedAt.ToDateTimeString(),
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
	return carbon.Parse(ce.ExpiresAt).IsPast()
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

	// result := strings.TrimSuffix(key, "_expires_at")
	// result = strings.TrimSuffix(result, "_hash")
	// result = strings.TrimSuffix(result, "_updated_at")

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
		c.Name = path.Base(utils.WorkingDir())
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
	ticker := time.NewTicker(30 * time.Second)
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

	// if c.IsBaseKey(key) && c.IsExpired(key) {
	// 	//c.purgeExpired()
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

	c.Db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(c.Name))
		cur := b.Cursor()

		primaryKeys := []string{}

		for k, _ := cur.First(); k != nil; k, _ = cur.Next() {
			if c.IsBaseKey(string(k)) {
				primaryKeys = append(primaryKeys, string(k))
			}
		}

		for _, key := range primaryKeys {
			if c.IsExpired(key) {
				fmt.Printf("Removing expired cache entry: %s\n", key)
				c.Remove(key)
			}
		}

		return nil
	})

}

// The `GetExpiresAt` function in the `Cache` struct is used to retrieve the expiration time of a cache
// entry with a given key.
func (c *Cache) GetExpiresAt(key string) *carbon.Carbon {
	value, found := c.Get(c.GetBaseKey(key))

	if !found {
		return nil
	}

	result := carbon.Parse(value.ExpiresAt)
	return &result
}

func (c *Cache) GetUpdatedAt(key string) *carbon.Carbon {
	value, found := c.Get(c.GetBaseKey(key))

	if !found {
		return nil
	}

	result := carbon.Parse(value.UpdatedAt)
	return &result
}

// The `IsExpired` function in the `Cache` struct is used to check if a cache entry with a given key
// has expired.
func (c *Cache) IsExpired(key string) bool {
	item, found := c.Get(c.GetBaseKey(key))
	if !found {
		return true
	}

	result := carbon.Parse(item.ExpiresAt)
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
		value := b.Get([]byte(key))
		found = value != nil

		return nil
	})

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
		baseKey := c.GetBaseKey(key)

		b.Delete([]byte(baseKey))
		// b.Delete([]byte(baseKey + "_hash"))
		// b.Delete([]byte(baseKey + "_expires_at"))
		// b.Delete([]byte(baseKey + "_updated_at"))

		return nil
	})
}
