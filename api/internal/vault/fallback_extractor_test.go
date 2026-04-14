package vault

import (
	"errors"
	"strings"
	"testing"
)

type fakeNamedExtractor struct {
	docType string
	text    string
	err     error
}

func (f fakeNamedExtractor) ProcessItem(_ Item, _ []byte) (string, string, error) {
	return f.docType, f.text, f.err
}

func TestFallbackMatrixPrimarySuccess(t *testing.T) {
	matrix := NewFallbackMatrixExtractor(
		map[string]FallbackStrategy{
			"pdf": {
				Primary: NamedExtractor{
					Name:      "gemini",
					Extractor: fakeNamedExtractor{docType: "presupuesto", text: "primary result"},
				},
				Secondary: NamedExtractor{
					Name:      "simulated",
					Extractor: fakeNamedExtractor{docType: "documento_pdf", text: "secondary result"},
				},
				RequireManualOn: true,
			},
		},
		FallbackStrategy{},
	)

	docType, extracted, err := matrix.ProcessItem(Item{MimeType: "application/pdf"}, []byte("x"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if docType != "presupuesto" {
		t.Fatalf("unexpected docType: %s", docType)
	}
	if extracted != "primary result" {
		t.Fatalf("unexpected extraction: %s", extracted)
	}
}

func TestFallbackMatrixSecondarySuccess(t *testing.T) {
	matrix := NewFallbackMatrixExtractor(
		map[string]FallbackStrategy{
			"pdf": {
				Primary: NamedExtractor{
					Name:      "gemini",
					Extractor: fakeNamedExtractor{err: errors.New("timeout")},
				},
				Secondary: NamedExtractor{
					Name:      "simulated",
					Extractor: fakeNamedExtractor{docType: "documento_pdf", text: "secondary result"},
				},
				RequireManualOn: true,
			},
		},
		FallbackStrategy{},
	)

	docType, extracted, err := matrix.ProcessItem(Item{MimeType: "application/pdf"}, []byte("x"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if docType != "documento_pdf" {
		t.Fatalf("unexpected docType: %s", docType)
	}
	if !strings.Contains(extracted, "fallback_provider=simulated") {
		t.Fatalf("expected fallback marker in text, got %q", extracted)
	}
}

func TestFallbackMatrixManualReviewRequired(t *testing.T) {
	matrix := NewFallbackMatrixExtractor(
		map[string]FallbackStrategy{
			"image": {
				Primary: NamedExtractor{
					Name:      "gemini",
					Extractor: fakeNamedExtractor{err: errors.New("network timeout")},
				},
				Secondary: NamedExtractor{
					Name:      "simulated",
					Extractor: fakeNamedExtractor{err: errors.New("unable to classify document type")},
				},
				RequireManualOn: true,
			},
		},
		FallbackStrategy{},
	)

	_, _, err := matrix.ProcessItem(Item{MimeType: "image/png"}, []byte("x"))
	if err == nil {
		t.Fatal("expected manual_review_required error")
	}
	if !strings.Contains(err.Error(), "manual_review_required") {
		t.Fatalf("expected manual review marker, got %v", err)
	}
}
