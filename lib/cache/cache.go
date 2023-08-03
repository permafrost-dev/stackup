package cache

import (
	"fmt"
	"os"
	"path"
	"path/filepath"

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
}

// The `Get` function in the `Cache` struct is used to retrieve the value of a cache entry with a given
// key. It takes a `key` parameter (string) and returns the corresponding value (string).
func (c *Cache) Get(key string) string {
	var result string

	// if !c.Has(key) {
	// 	return ""
	// }

	// if c.IsExpired(key) {
	// 	c.purgeExpired()
	// 	return ""
	// }

	c.Db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(c.Name))
		bytes := b.Get([]byte(key))
		result = string(bytes)
		return nil
	})

	return result
}

// The `purgeExpired` function in the `Cache` struct is used to remove any cache entries that have
// expired. It iterates through all the keys in the cache bucket and checks if each key has expired
// using the `IsExpired` function. If a key is expired, it is deleted from the cache bucket. This
// function ensures that expired cache entries are automatically removed from the cache to free up
// space and maintain cache integrity.
func (c *Cache) purgeExpired() {
	return
	c.Db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(c.Name))
		cur := b.Cursor()

		for k, _ := cur.First(); k != nil; k, _ = cur.Next() {
			if c.IsExpired(string(k)) {
				b.Delete(k)
			}
		}

		return nil
	})
}

// The `GetExpiresAt` function in the `Cache` struct is used to retrieve the expiration time of a cache
// entry with a given key.
func (c *Cache) GetExpiresAt(key string) *carbon.Carbon {
	value := c.Get(key + "_expires_at")
	time := carbon.Parse(value, "America/New_York")
	return &time
}

// The `GetHash` function in the `Cache` struct is used to retrieve the hash value stored in the cache
// for a given key. It takes a `key` parameter (string) and returns the corresponding hash value
// (string). It calls the `Get` function to retrieve the value of the cache entry with the given key
// appended with "_hash".
func (c *Cache) GetHash(key string) string {
	value := c.Get(key + "_hash")
	return value
}

// The `IsExpired` function in the `Cache` struct is used to check if a cache entry with a given key
// has expired.
func (c *Cache) IsExpired(key string) bool {
	expiresAt := c.GetExpiresAt(key)
	if expiresAt == nil {
		return false
	}
	return expiresAt.IsPast()
}

// The `HashMatches` function in the `Cache` struct is used to check if the hash value stored in the
// cache for a given key matches a provided hash value.
func (c *Cache) HashMatches(key string, hash string) bool {
	return c.GetHash(key) == hash
}

// The `Set` function in the `Cache` struct is used to set a cache entry with a given key and value. It
// takes three parameters: `key` (string), `value` (any), and `ttlMinutes` (int).
func (c *Cache) Set(key string, value any, ttlMinutes int) {
	c.Db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(c.Name))

		if value == nil {
			b.Delete([]byte(key))
			b.Delete([]byte(key + "_expires_at"))
			return nil
		}

		err := b.Put([]byte(key), []byte(value.(string)))
		expiresAt := carbon.Timestamp{carbon.Now().AddMinutes(ttlMinutes)}
		b.Put([]byte(key+"_expires_at"), []byte(expiresAt.String()))

		hash := utils.CalculateSHA256Hash(value.(string))
		b.Put([]byte(key+"_hash"), []byte(hash))

		return err
	})
}

// The `Has` function in the `Cache` struct is used to check if a cache entry with a given key exists
// and is not expired. It calls the `Get` function to retrieve the value of the cache entry with the
// given key and checks if the value is not empty (`c.Get(key) != ""`) and if the cache entry is not
// expired (`!c.IsExpired(key)`). If both conditions are true, it returns `true`, indicating that the
// cache entry exists and is valid. Otherwise, it returns `false`.
func (c *Cache) Has(key string) bool {
	return c.Get(key) != "" && !c.IsExpired(key)
}

// The `Remove` function in the `Cache` struct is used to remove a cache entry with a given key. It
// calls the `Set` function with a `value` parameter of `nil` and a `ttlMinutes` parameter of `0`. This
// effectively sets the cache entry to be empty and expired, effectively removing it from the cache.
func (c *Cache) Remove(key string) {
	c.Set(key, nil, 0)
}
