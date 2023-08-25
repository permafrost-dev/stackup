package cache

import (
	"encoding/base64"
	"fmt"

	carbon "github.com/golang-module/carbon/v2"
)

type CacheEntry struct {
	Value       string `json:"value"`
	Hash        string `json:"hash"`
	Algorithm   string `json:"algorithm"`
	ExpiresAtTs carbon.Carbon
	UpdatedAtTs carbon.Carbon
	ExpiresAt   string `json:"expires_at"`
	UpdatedAt   string `json:"updated_at"`
}

func (ce *CacheEntry) IsExpired() bool {
	if ce.ExpiresAt == "" {
		return true
	}

	return ce.ExpiresAtTs.IsPast()

	//return carbon.Parse(ce.ExpiresAt).IsPast()
}

func (ce *CacheEntry) EncodeValue() {
	ce.Value = base64.StdEncoding.EncodeToString([]byte(ce.Value))
}

func (ce *CacheEntry) DecodeValue() {
	fmt.Printf(" [trace] ==> CacheEntry.DecodeValue()\n")
	defer fmt.Printf(" [trace] <== CacheEntry.DecodeValue()\n")

	if ce.Value == "" {
		return
	}

	decoded, err := base64.StdEncoding.DecodeString(ce.Value)
	if err != nil {
		return
	}

	ce.Value = string(decoded)
}

func (ce *CacheEntry) UpdateTimestampsFromStrings() error {
	//func (ce *CacheEntry) UnmarshalJson(data []byte) error {
	//if ce.ExpiresAt != "" {
	if parsed := carbon.Parse(ce.ExpiresAt); parsed.Error == nil {
		ce.ExpiresAtTs = parsed
	} else {
		return parsed.Error
	}

	if parsed := carbon.Parse(ce.UpdatedAt); parsed.Error == nil {
		ce.UpdatedAtTs = parsed
	} else {
		return parsed.Error
	}

	return nil
}

func (ce *CacheEntry) UpdateTimestampsFromObjects() {
	if ce.ExpiresAtTs.IsValid() {
		ce.ExpiresAt = ce.ExpiresAtTs.ToIso8601String()
	}

	if ce.UpdatedAtTs.IsValid() {
		ce.UpdatedAt = ce.UpdatedAtTs.ToIso8601String()
	}
}
