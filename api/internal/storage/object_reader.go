package storage

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

var ErrObjectReaderNotConfigured = fmt.Errorf("object reader not configured")

type ObjectReader interface {
	ReadObject(storagePath string) ([]byte, error)
}

type NoopObjectReader struct{}

func (NoopObjectReader) ReadObject(_ string) ([]byte, error) {
	return nil, ErrObjectReaderNotConfigured
}

type SupabaseObjectReader struct {
	BaseURL        string
	ServiceRoleKey string
	Bucket         string
	HTTPClient     *http.Client
}

func (s SupabaseObjectReader) ReadObject(storagePath string) ([]byte, error) {
	if s.BaseURL == "" || s.ServiceRoleKey == "" || s.Bucket == "" {
		return nil, ErrObjectReaderNotConfigured
	}

	client := s.HTTPClient
	if client == nil {
		client = &http.Client{Timeout: 20 * time.Second}
	}

	base := strings.TrimRight(s.BaseURL, "/")
	url := fmt.Sprintf("%s/storage/v1/object/%s/%s", base, s.Bucket, storagePath)

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+s.ServiceRoleKey)
	req.Header.Set("apikey", s.ServiceRoleKey)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("failed to read object with status %d", resp.StatusCode)
	}

	return io.ReadAll(resp.Body)
}
