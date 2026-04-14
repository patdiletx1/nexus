package chilecompra

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"nexus/api/internal/tenders"
)

var ErrNotConfigured = fmt.Errorf("chilecompra client not configured")

type Client interface {
	FetchTenders(ctx context.Context, since *time.Time, limit int) ([]tenders.Tender, error)
}

type NoopClient struct{}

func (NoopClient) FetchTenders(_ context.Context, _ *time.Time, _ int) ([]tenders.Tender, error) {
	return nil, ErrNotConfigured
}

type APIClient struct {
	BaseURL     string
	APIKey      string
	TendersPath string
	HTTPClient  *http.Client
}

func (c APIClient) FetchTenders(ctx context.Context, since *time.Time, limit int) ([]tenders.Tender, error) {
	if c.BaseURL == "" || c.APIKey == "" {
		return nil, ErrNotConfigured
	}
	if limit <= 0 {
		limit = 50
	}
	if limit > 200 {
		limit = 200
	}

	client := c.HTTPClient
	if client == nil {
		client = &http.Client{Timeout: 20 * time.Second}
	}

	path := strings.TrimSpace(c.TendersPath)
	if path == "" {
		path = "/servicios/v1/publico/licitaciones.json"
	}
	base := strings.TrimRight(c.BaseURL, "/")
	u, err := url.Parse(base + path)
	if err != nil {
		return nil, err
	}

	query := u.Query()
	query.Set("limit", strconv.Itoa(limit))
	if since != nil {
		query.Set("since", since.UTC().Format(time.RFC3339))
	}
	u.RawQuery = query.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.APIKey)
	req.Header.Set("apikey", c.APIKey)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("chilecompra request failed with status %d", resp.StatusCode)
	}

	var raw any
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, err
	}

	return parseTenders(raw), nil
}

func parseTenders(raw any) []tenders.Tender {
	items := make([]map[string]any, 0)
	switch typed := raw.(type) {
	case []any:
		for _, value := range typed {
			if item, ok := value.(map[string]any); ok {
				items = append(items, item)
			}
		}
	case map[string]any:
		if values, ok := typed["data"].([]any); ok {
			for _, value := range values {
				if item, ok := value.(map[string]any); ok {
					items = append(items, item)
				}
			}
		} else if values, ok := typed["Listado"].([]any); ok {
			for _, value := range values {
				if item, ok := value.(map[string]any); ok {
					items = append(items, item)
				}
			}
		}
	}

	out := make([]tenders.Tender, 0, len(items))
	for _, item := range items {
		externalID := firstString(item, "id", "codigo", "CodigoExterno")
		title := firstString(item, "title", "nombre", "Nombre")
		if externalID == "" || title == "" {
			continue
		}

		tender := tenders.Tender{
			ExternalID:    externalID,
			Title:         title,
			Description:   firstString(item, "description", "descripcion", "Descripcion"),
			Region:        firstString(item, "region", "Region"),
			ClosingAt:     parseTimePointer(firstString(item, "closing_at", "fecha_cierre", "FechaCierre")),
			PublishedAt:   parseTimePointer(firstString(item, "published_at", "fecha_publicacion", "FechaPublicacion")),
			Source:        "chilecompra",
			SourcePayload: item,
		}
		out = append(out, tender)
	}
	return out
}

func firstString(item map[string]any, keys ...string) string {
	for _, key := range keys {
		value, ok := item[key]
		if !ok {
			continue
		}
		if asString, ok := value.(string); ok {
			trimmed := strings.TrimSpace(asString)
			if trimmed != "" {
				return trimmed
			}
		}
	}
	return ""
}

func parseTimePointer(value string) *time.Time {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	layouts := []string{time.RFC3339, "2006-01-02 15:04:05", "2006-01-02"}
	for _, layout := range layouts {
		if parsed, err := time.Parse(layout, value); err == nil {
			parsed = parsed.UTC()
			return &parsed
		}
	}
	return nil
}
