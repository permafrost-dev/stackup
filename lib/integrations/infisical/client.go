package infisical

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"strings"
)

type Client struct {
	APIEndpoint string
	HTTPClient  *http.Client
}

func NewClient(endpoint string) *Client {
	return &Client{
		APIEndpoint: endpoint,
		HTTPClient:  &http.Client{},
	}
}

type TokenData struct {
	EncryptedKey string `json:"encryptedKey"`
	Iv           string `json:"iv"`
	Tag          string `json:"tag"`
	Name         string `json:"name"`
}

type Secret struct {
	Name                  string `json:"name"`
	Value                 string `json:"value"`
	Type                  string `json:"type"`
	WorkspaceID           string `json:"workspaceId"`
	Environment           string `json:"environment"`
	Path                  string `json:"path"`
	SecretKeyCiphertext   string `json:"secretKeyCiphertext"`
	SecretKeyIV           string `json:"secretKeyIV"`
	SecretKeyTag          string `json:"secretKeyTag"`
	SecretKeyHash         string `json:"secretKeyHash"`
	SecretValueCiphertext string `json:"secretValueCiphertext"`
	SecretValueIV         string `json:"secretValueIV"`
	SecretValueTag        string `json:"secretValueTag"`
	SecretValueHash       string `json:"secretValueHash"`
}

type CreateSecretRequest struct {
	SecretName            string `json:"secretName"`
	WorkspaceID           string `json:"workspaceId"`
	Environment           string `json:"environment"`
	Type                  string `json:"type"`
	SecretKeyCiphertext   string `json:"secretKeyCiphertext"`
	SecretKeyIV           string `json:"secretKeyIV"`
	SecretKeyTag          string `json:"secretKeyTag"`
	SecretValueCiphertext string `json:"secretValueCiphertext"`
	SecretValueIV         string `json:"secretValueIV"`
	SecretValueTag        string `json:"secretValueTag"`
	Path                  string `json:"path"`
}

type UpdateSecretRequest struct {
	SecretName            string `json:"secretName"`
	WorkspaceID           string `json:"workspaceId"`
	Environment           string `json:"environment"`
	Type                  string `json:"type"`
	SecretValueCiphertext string `json:"secretValueCiphertext"`
	SecretValueIV         string `json:"secretValueIV"`
	SecretValueTag        string `json:"secretValueTag"`
}

func (c *Client) makeRequest(method, endpoint string, payload interface{}) (*http.Response, error) {
	url := c.APIEndpoint + endpoint
	var req *http.Request
	var err error

	if method == http.MethodGet || method == http.MethodDelete {
		req, err = http.NewRequest(method, url, nil)
	} else {
		jsonPayload, _ := json.Marshal(payload)
		req, err = http.NewRequest(method, url, bytes.NewBuffer(jsonPayload))
		req.Header.Set("Content-Type", "application/json")
	}

	if err != nil {
		return nil, err
	}

	token := os.Getenv("INFISICAL_API_TOKEN")
	apiKey := os.Getenv("INFISICAL_API_KEY")

	req.Header.Set("User-Agent", "StackUp-cli/v1")

	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	if apiKey != "" {
		req.Header.Del("Authorization")
		req.Header.Set("X-API-KEY", apiKey)
	}

	if apiKey == "" && token == "" {
		return nil, errors.New("INFISICAL_API_KEY or INFISICAL_API_TOKEN environment variable must be set")
	}

	return c.HTTPClient.Do(req)
}

func decrypt(ciphertext, iv, tag, secret string) (string, error) {
	cipherTextBytes, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", err
	}

	ivBytes, err := base64.StdEncoding.DecodeString(iv)
	if err != nil {
		return "", err
	}

	tagBytes, err := base64.StdEncoding.DecodeString(tag)
	if err != nil {
		return "", err
	}

	secretBytes := []byte(secret)

	block, err := aes.NewCipher(secretBytes)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCMWithNonceSize(block, len(ivBytes))
	if err != nil {
		return "", err
	}

	combinedCipherText := append(cipherTextBytes, tagBytes...)
	plainText, err := gcm.Open(nil, ivBytes, combinedCipherText, nil)
	if err != nil {
		return "", errors.New("failed to decrypt")
	}

	return string(plainText), nil
}

func (c *Client) GetSecrets(workspaceId, environment, path string, includeImports bool) ([]Secret, error) {
	endpoint := "/api/v3/secrets?workspaceId=" + workspaceId + "&environment=" + environment + "&path=" + path
	if includeImports {
		endpoint += "&include_imports=true"
	}

	resp, err := c.makeRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var secretsResponse struct {
		Secrets []Secret `json:"secrets"`
	}

	// body := make([]byte, 81920)
	// resp.Body.Read(body)
	// fmt.Printf("resp.Body: %v\n", string(body))

	err = json.NewDecoder(resp.Body).Decode(&secretsResponse)
	if err != nil {
		return nil, err
	}

	result := secretsResponse.Secrets
	token := os.Getenv("INFISICAL_API_TOKEN")

	lastIndex := strings.LastIndex(token, ".")
	tokenSecret := ""

	if lastIndex != -1 && lastIndex+1 < len(token) {
		tokenSecret = token[lastIndex+1:]
	}

	tokenData, _ := c.GetTokenData(workspaceId, environment)

	secretKey, err := decrypt(tokenData.EncryptedKey, tokenData.Iv, tokenData.Tag, tokenSecret)
	if err != nil {
		return nil, err
	}

	for _, secret := range result {
		if secret.SecretKeyCiphertext != "" {
			plainTextKey, err := decrypt(secret.SecretKeyCiphertext, secret.SecretKeyIV, secret.SecretKeyTag, secretKey)
			if err != nil {
				return nil, err
			}

			plainTextValue, err := decrypt(secret.SecretValueCiphertext, secret.SecretValueIV, secret.SecretValueTag, secretKey)
			if err != nil {
				return nil, err
			}

			secret.Name = plainTextKey
			secret.Value = plainTextValue
		}
	}

	return result, nil
}

func (c *Client) GetTokenData(workspaceId, environment string) (TokenData, error) {
	endpoint := "/api/v2/service-token"

	resp, err := c.makeRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return TokenData{}, err
	}
	defer resp.Body.Close()

	var secretResponse TokenData

	err = json.NewDecoder(resp.Body).Decode(&secretResponse)
	if err != nil {
		return TokenData{}, err
	}

	return secretResponse, nil
}

func (c *Client) GetSecret(workspaceId, environment, secretName, path string) (Secret, error) {
	endpoint := "/api/v3/secrets/" + secretName + "?workspaceId=" + workspaceId + "&environment=" + environment + "&path=" + path

	resp, err := c.makeRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return Secret{}, err
	}
	defer resp.Body.Close()

	var secretResponse struct {
		Secret Secret `json:"secret"`
	}
	err = json.NewDecoder(resp.Body).Decode(&secretResponse)
	if err != nil {
		return Secret{}, err
	}

	return secretResponse.Secret, nil
}

func (c *Client) CreateSecret(secret CreateSecretRequest) (Secret, error) {
	endpoint := "/api/v3/secrets/" + secret.SecretName

	resp, err := c.makeRequest(http.MethodPost, endpoint, secret)
	if err != nil {
		return Secret{}, err
	}
	defer resp.Body.Close()

	var secretResponse struct {
		Secret Secret `json:"secret"`
	}
	err = json.NewDecoder(resp.Body).Decode(&secretResponse)
	if err != nil {
		return Secret{}, err
	}

	return secretResponse.Secret, nil
}

func (c *Client) UpdateSecret(secret UpdateSecretRequest) (Secret, error) {
	endpoint := "/api/v3/secrets/" + secret.SecretName

	resp, err := c.makeRequest(http.MethodPatch, endpoint, secret)
	if err != nil {
		return Secret{}, err
	}
	defer resp.Body.Close()

	var secretResponse struct {
		Secret Secret `json:"secret"`
	}
	err = json.NewDecoder(resp.Body).Decode(&secretResponse)
	if err != nil {
		return Secret{}, err
	}

	return secretResponse.Secret, nil
}

func (c *Client) DeleteSecret(workspaceId, environment, secretName, path string) error {
	endpoint := "/api/v3/secrets/" + secretName + "?workspaceId=" + workspaceId + "&environment=" + environment + "&path=" + path

	resp, err := c.makeRequest(http.MethodDelete, endpoint, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return errors.New("Failed to delete secret")
	}

	return nil
}
