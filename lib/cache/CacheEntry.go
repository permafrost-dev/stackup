package cache

import (
	"encoding/base64"

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
	if ce.ExpiresAt == "" {
		return true
	}

	return carbon.Parse(ce.ExpiresAt).IsPast()
}

func (ce *CacheEntry) EncodeValue() {
	ce.Value = base64.StdEncoding.EncodeToString([]byte(ce.Value))
}

func (ce *CacheEntry) DecodeValue() {
	if ce.Value == "" {
		return
	}

	decoded, err := base64.StdEncoding.DecodeString(ce.Value)
	if err != nil {
		return
	}

	ce.Value = string(decoded)
}
