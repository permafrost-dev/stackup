package cache

import (
	carbon "github.com/golang-module/carbon/v2"
)

type CacheEntry struct {
	Key       string
	Value     string
	Hash      string
	Algorithm string
	ExpiresAt string
	UpdatedAt string
}

func (ce *CacheEntry) IsExpired() bool {
	return carbon.Parse(ce.ExpiresAt).IsPast()
}
