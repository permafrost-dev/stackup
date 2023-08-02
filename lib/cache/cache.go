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
	Db          *bolt.DB
	Enabled     bool
	ProjectName string
	Path        string
	Filename    string
}

func CreateCache() *Cache {
	result := Cache{Enabled: false}
	result.Init()

	return &result
}

func EnsureConfigDirExists(dirName string) (string, error) {
	// Get the user's home directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	// Append the directory name to the home directory
	configDir := filepath.Join(homeDir, dirName)

	// Ensure the directory exists
	err = os.MkdirAll(configDir, 0700)
	if err != nil {
		return "", err
	}

	return configDir, nil
}

func (c *Cache) Init() {
	if c.Db != nil {
		return
	}

	c.Path, _ = EnsureConfigDirExists(".stackup")
	c.Filename = filepath.Join(c.Path, "stackup.db")
	c.ProjectName = path.Base(utils.WorkingDir())

	db, err := bolt.Open(c.Filename, 0600, bolt.DefaultOptions)
	if err != nil {
		c.Enabled = false
		c.Db = nil
		return
	}

	c.Db = db

	// create a new project bucket if it doesn't exist
	c.Db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(c.ProjectName))
		if err != nil {
			return fmt.Errorf("create bucket: %s", err)
		}
		return nil
	})

	c.Enabled = true
}

func (c *Cache) Get(key string) string {
	var result string
	// expiresAt := c.GetExpiresAt(key + "_expires_at")

	// if expiresAt != nil && expiresAt.IsPast() {
	// 	return nil
	// }

	c.Db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(c.ProjectName))
		bytes := b.Get([]byte(key))
		result = string(bytes)
		return nil
	})

	return result
}

func (c *Cache) GetExpiresAt(key string) *carbon.Carbon {
	value := c.Get(key + "_expires_at")
	time := carbon.Parse(value, "America/New_York")
	return &time
}

func (c *Cache) GetHash(key string) string {
	value := c.Get(key + "_hash")
	return value
}

func (c *Cache) IsExpired(key string) bool {
	expiresAt := c.GetExpiresAt(key)
	if expiresAt == nil {
		return false
	}
	return expiresAt.IsPast()
}

func (c *Cache) Set(key string, value any, ttlMinutes int) {
	c.Db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(c.ProjectName))

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

func (c *Cache) Has(key string) bool {
	return c.Get(key) != ""
}
