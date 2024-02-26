package infisical

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetSecrets(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		secrets := []Secret{
			{Name: "test-secret", Value: "test-value"},
		}
		json.NewEncoder(w).Encode(map[string][]Secret{"secrets": secrets})
	}))
	defer server.Close()

	client := NewClient(server.URL)
	secrets, err := client.GetSecrets("workspaceId", "environment", "path", false)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(secrets) != 1 {
		t.Fatalf("Expected 1 secret, got %d", len(secrets))
	}
}

func TestGetSecret(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		secret := Secret{Name: "test-secret", Value: "test-value"}
		json.NewEncoder(w).Encode(map[string]Secret{"secret": secret})
	}))
	defer server.Close()

	client := NewClient(server.URL)
	secret, err := client.GetSecret("workspaceId", "environment", "test-secret", "path")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if secret.Name != "test-secret" {
		t.Fatalf("Expected secret name to be test-secret, got %s", secret.Name)
	}

	if secret.Value != "test-value" {
		t.Fatalf("Expected secret value to be test-value, got %s", secret.Value)
	}
}

func TestCreateSecret(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		secret := Secret{Name: "test-secret", Value: "test-value"}
		json.NewEncoder(w).Encode(map[string]Secret{"secret": secret})
	}))
	defer server.Close()

	client := NewClient(server.URL)
	req := CreateSecretRequest{SecretName: "test-secret"}
	secret, err := client.CreateSecret(req)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if secret.Name != "test-secret" {
		t.Fatalf("Expected secret name to be test-secret, got %s", secret.Name)
	}
}

func TestUpdateSecret(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		secret := Secret{Name: "test-secret", Value: "updated-value"}
		json.NewEncoder(w).Encode(map[string]Secret{"secret": secret})
	}))
	defer server.Close()

	client := NewClient(server.URL)
	req := UpdateSecretRequest{SecretName: "test-secret"}
	secret, err := client.UpdateSecret(req)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if secret.Value != "updated-value" {
		t.Fatalf("Expected secret value to be updated-value, got %s", secret.Value)
	}
}

func TestDeleteSecret(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClient(server.URL)
	err := client.DeleteSecret("workspaceId", "environment", "test-secret", "path")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
}

func TestMakeRequest(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}))
	defer server.Close()

	client := NewClient(server.URL)
	resp, err := client.makeRequest(http.MethodGet, "/", nil)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	defer resp.Body.Close()

	buf := new(bytes.Buffer)
	buf.ReadFrom(resp.Body)
	if buf.String() != "OK" {
		t.Fatalf("Expected response body to be OK, got %s", buf.String())
	}
}
