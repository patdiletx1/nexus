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

func TestMockClientFetchTendersRespectsLimit(t *testing.T) {
	client := MockClient{}
	items, err := client.FetchTenders(nil, nil, 2)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(items))
	}
	if items[0].ExternalID == "" || items[1].ExternalID == "" {
		t.Fatalf("expected mock tenders with external ids, got %+v", items)
	}
}
