package infisical

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
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

type Secret struct {
	Name        string `json:"name"`
	Value       string `json:"value"`
	Type        string `json:"type"`
	WorkspaceID string `json:"workspaceId"`
	Environment string `json:"environment"`
	Path        string `json:"path"`
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

	return c.HTTPClient.Do(req)
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
	err = json.NewDecoder(resp.Body).Decode(&secretsResponse)
	if err != nil {
		return nil, err
	}

	return secretsResponse.Secrets, nil
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
