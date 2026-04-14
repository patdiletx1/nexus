package vault

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

const minDatasetAccuracy = 0.85

type datasetCase struct {
	Name            string `json:"name"`
	FileName        string `json:"file_name"`
	MimeType        string `json:"mime_type"`
	ExpectedDocType string `json:"expected_doc_type"`
}

func TestDocumentTypeDatasetAccuracy(t *testing.T) {
	rawDataset, err := os.ReadFile(filepath.Join("..", "..", "testdata", "vault_extraction_dataset.json"))
	if err != nil {
		t.Fatalf("failed to read dataset: %v", err)
	}

	var cases []datasetCase
	if err := json.Unmarshal(rawDataset, &cases); err != nil {
		t.Fatalf("failed to parse dataset: %v", err)
	}
	if len(cases) == 0 {
		t.Fatal("dataset cannot be empty")
	}

	hits := 0
	for _, tc := range cases {
		predicted := detectDocumentType(tc.FileName, tc.MimeType)
		if predicted == tc.ExpectedDocType {
			hits++
			continue
		}

		t.Logf("mismatch case=%q expected=%q predicted=%q", tc.Name, tc.ExpectedDocType, predicted)
	}

	accuracy := float64(hits) / float64(len(cases))
	t.Logf("dataset_accuracy=%.2f hits=%d total=%d threshold=%.2f", accuracy, hits, len(cases), minDatasetAccuracy)
	if accuracy < minDatasetAccuracy {
		t.Fatalf("document type accuracy below threshold: got %.2f expected >= %.2f", accuracy, minDatasetAccuracy)
	}
}

func TestClassifyProcessingError(t *testing.T) {
	tests := []struct {
		name              string
		stage             string
		err               error
		expectedCategory  string
		expectedRetryable bool
	}{
		{
			name:              "timeout extraction",
			stage:             "extract_content",
			err:               os.ErrDeadlineExceeded,
			expectedCategory:  "timeout",
			expectedRetryable: true,
		},
		{
			name:              "unsupported document",
			stage:             "extract_content",
			err:               errString("unable to classify document type"),
			expectedCategory:  "unsupported_document",
			expectedRetryable: false,
		},
		{
			name:              "configuration missing",
			stage:             "read_object",
			err:               errString("service not configured"),
			expectedCategory:  "configuration",
			expectedRetryable: false,
		},
		{
			name:              "transient upstream",
			stage:             "read_object",
			err:               errString("connection reset by peer"),
			expectedCategory:  "transient_upstream",
			expectedRetryable: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			details := ClassifyProcessingError(tc.stage, tc.err)
			if details.Category != tc.expectedCategory {
				t.Fatalf("expected category %s, got %s", tc.expectedCategory, details.Category)
			}
			if details.Retryable != tc.expectedRetryable {
				t.Fatalf("expected retryable %t, got %t", tc.expectedRetryable, details.Retryable)
			}
			if details.Stage != tc.stage {
				t.Fatalf("expected stage %s, got %s", tc.stage, details.Stage)
			}
		})
	}
}

type errString string

func (e errString) Error() string {
	return string(e)
}
