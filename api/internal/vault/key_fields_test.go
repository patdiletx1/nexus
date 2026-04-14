package vault

import "testing"

func TestExtractKeyFields(t *testing.T) {
	text := `
Resumen propuesta
Razon social: Acme Industrial SpA
Monto total: $12.345.678,90
Fecha cierre: 14/04/2026
`

	fields := ExtractKeyFields(text)
	if fields.CompanyName != "Acme Industrial SpA" {
		t.Fatalf("unexpected company name: %q", fields.CompanyName)
	}
	if fields.Amount != "$12.345.678,90" {
		t.Fatalf("unexpected amount: %q", fields.Amount)
	}
	if fields.Date != "2026-04-14" {
		t.Fatalf("unexpected normalized date: %q", fields.Date)
	}
	if fields.FoundCount() != 3 {
		t.Fatalf("expected 3 found fields, got %d", fields.FoundCount())
	}
}

func TestMissingRequiredFields(t *testing.T) {
	fields := ExtractKeyFields("Documento sin campos relevantes")
	missing := fields.MissingRequired()
	if len(missing) != 3 {
		t.Fatalf("expected 3 missing fields, got %d", len(missing))
	}
}
