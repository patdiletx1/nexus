package chilecompra

import "testing"

func TestParseTendersFromDataEnvelope(t *testing.T) {
	raw := map[string]any{
		"data": []any{
			map[string]any{
				"id":          "123",
				"title":       "Licitacion A",
				"description": "Servicio tecnico",
				"region":      "Metropolitana",
				"closing_at":  "2026-05-01T10:00:00Z",
			},
		},
	}

	items := parseTenders(raw)
	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}
	if items[0].ExternalID != "123" || items[0].Title != "Licitacion A" {
		t.Fatalf("unexpected parsed tender: %+v", items[0])
	}
}
