package vault

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const minKeyFieldsCoverage = 0.80

type keyFieldsDatasetCase struct {
	Name                string `json:"name"`
	Text                string `json:"text"`
	ExpectedAmount      string `json:"expected_amount"`
	ExpectedDate        string `json:"expected_date"`
	ExpectedCompanyName string `json:"expected_company_name"`
}

func TestKeyFieldsCoverageDataset(t *testing.T) {
	rawDataset, err := os.ReadFile(filepath.Join("..", "..", "testdata", "vault_key_fields_dataset.json"))
	if err != nil {
		t.Fatalf("failed to read dataset: %v", err)
	}

	var cases []keyFieldsDatasetCase
	if err := json.Unmarshal(rawDataset, &cases); err != nil {
		t.Fatalf("failed to parse dataset: %v", err)
	}
	if len(cases) == 0 {
		t.Fatal("dataset cannot be empty")
	}

	totalExpected := 0
	totalHits := 0
	for _, tc := range cases {
		got := ExtractKeyFields(tc.Text)
		totalExpected += countNonEmpty(tc.ExpectedAmount, tc.ExpectedDate, tc.ExpectedCompanyName)
		totalHits += matchField(got.Amount, tc.ExpectedAmount)
		totalHits += matchField(got.Date, tc.ExpectedDate)
		totalHits += matchField(got.CompanyName, tc.ExpectedCompanyName)
	}

	if totalExpected == 0 {
		t.Fatal("dataset must have at least one expected field")
	}
	coverage := float64(totalHits) / float64(totalExpected)
	t.Logf("key_fields_coverage=%.2f hits=%d total_expected=%d threshold=%.2f", coverage, totalHits, totalExpected, minKeyFieldsCoverage)
	if coverage < minKeyFieldsCoverage {
		t.Fatalf("key field coverage below threshold: got %.2f expected >= %.2f", coverage, minKeyFieldsCoverage)
	}
}

func countNonEmpty(values ...string) int {
	total := 0
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			total++
		}
	}
	return total
}

func matchField(got, expected string) int {
	if strings.TrimSpace(expected) == "" {
		return 0
	}
	if strings.TrimSpace(got) == strings.TrimSpace(expected) {
		return 1
	}
	return 0
}
