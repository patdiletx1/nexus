package storage

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type SignedUpload struct {
	URL       string
	ExpiresAt time.Time
}

type Signer interface {
	SignUpload(storagePath string, expiresInSeconds int) (SignedUpload, error)
}

type PlaceholderSigner struct{}

func (PlaceholderSigner) SignUpload(storagePath string, expiresInSeconds int) (SignedUpload, error) {
	now := time.Now().UTC()
	return SignedUpload{
		URL:       fmt.Sprintf("/v1/vault/upload/direct/%s", storagePath),
		ExpiresAt: now.Add(time.Duration(expiresInSeconds) * time.Second),
	}, nil
}

type SupabaseSigner struct {
	BaseURL        string
	ServiceRoleKey string
	Bucket         string
	HTTPClient     *http.Client
}

func (s SupabaseSigner) SignUpload(storagePath string, expiresInSeconds int) (SignedUpload, error) {
	if s.BaseURL == "" || s.ServiceRoleKey == "" || s.Bucket == "" {
		return SignedUpload{}, fmt.Errorf("supabase signer misconfigured")
	}

	client := s.HTTPClient
	if client == nil {
		client = &http.Client{Timeout: 10 * time.Second}
	}

	base := strings.TrimRight(s.BaseURL, "/")
	path := url.PathEscape(storagePath)
	endpoint := fmt.Sprintf("%s/storage/v1/object/sign/%s/%s", base, s.Bucket, path)

	body := map[string]any{"expiresIn": expiresInSeconds}
	payload, _ := json.Marshal(body)
	req, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewReader(payload))
	if err != nil {
		return SignedUpload{}, err
	}

	req.Header.Set("Authorization", "Bearer "+s.ServiceRoleKey)
	req.Header.Set("apikey", s.ServiceRoleKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return SignedUpload{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return SignedUpload{}, fmt.Errorf("failed to sign upload url with status %d", resp.StatusCode)
	}

	var out struct {
		SignedURL string `json:"signedURL"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return SignedUpload{}, err
	}
	if out.SignedURL == "" {
		return SignedUpload{}, fmt.Errorf("missing signed url")
	}

	now := time.Now().UTC()
	return SignedUpload{
		URL:       base + out.SignedURL,
		ExpiresAt: now.Add(time.Duration(expiresInSeconds) * time.Second),
	}, nil
}
