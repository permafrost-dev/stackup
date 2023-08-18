package app

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"

	"github.com/stackup-app/stackup/lib/utils"
	"gopkg.in/yaml.v2"
)

type RemoteTemplateIndex struct {
	Loaded    bool
	Templates []*RemoteTemplate `yaml:"templates"`
}

type RemoteTemplate struct {
	Name      string `yaml:"name"`
	Location  string `yaml:"location"`
	Checksum  string `yaml:"checksum"`
	Algorithm string `yaml:"algorithm"`
}

func LoadRemoteTemplateIndex(url string) (*RemoteTemplateIndex, error) {
	body, err := utils.GetUrlContents(url)

	var index RemoteTemplateIndex
	err = yaml.Unmarshal([]byte(body), &index)
	if err != nil {
		return nil, err
	}

	return &index, nil
}

func (index *RemoteTemplateIndex) GetTemplate(name string) *RemoteTemplate {
	for _, template := range index.Templates {
		if template.Location == name {
			return template
		}
	}

	return nil
}

func (t *RemoteTemplate) GetContents() ([]byte, error) {
	// Send an HTTP GET request to the location URL
	resp, err := http.Get(t.Location)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Read the response body into a byte slice
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}

func (t *RemoteTemplate) ValidateChecksum(contents string) (bool, error) {
	var hash []byte
	switch t.Algorithm {
	case "sha256":
		h := sha256.New()
		h.Write([]byte(contents))
		hash = h.Sum(nil)
		break
	case "sha512":
		h := sha512.New()
		h.Write([]byte(contents))
		hash = h.Sum(nil)
		break
	default:
		return false, fmt.Errorf("unsupported algorithm: %s", t.Algorithm)
	}

	// fmt.Printf("hash: %x\n", hash)
	// fmt.Printf("checksum: %s\n", t.Checksum)

	checksumBytes, err := hex.DecodeString(t.Checksum)
	if err != nil {
		return false, err
	}
	if !hmac.Equal(hash, checksumBytes) {
		return false, nil
	}

	return true, nil
}
