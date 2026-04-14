package vault

import (
	"regexp"
	"strings"
	"time"
)

var (
	currencyAmountPattern = regexp.MustCompile(`(?i)\$\s*[0-9]{1,3}(?:\.[0-9]{3})*(?:,[0-9]{2})?`)
	isoDatePattern        = regexp.MustCompile(`\b\d{4}-\d{2}-\d{2}\b`)
	chileDatePattern      = regexp.MustCompile(`\b\d{2}/\d{2}/\d{4}\b`)
	companyLinePattern    = regexp.MustCompile(`(?i)(raz[oó]n social|empresa|proveedor)\s*:\s*([^\n\r]+)`)
)

type KeyFields struct {
	Amount      string `json:"amount,omitempty"`
	Date        string `json:"date,omitempty"`
	CompanyName string `json:"company_name,omitempty"`
}

func ExtractKeyFields(extractedText string) KeyFields {
	text := strings.TrimSpace(extractedText)
	if text == "" {
		return KeyFields{}
	}

	var out KeyFields
	if amount := currencyAmountPattern.FindString(text); amount != "" {
		out.Amount = strings.TrimSpace(amount)
	}

	if dateRaw := findDate(text); dateRaw != "" {
		out.Date = dateRaw
	}

	if match := companyLinePattern.FindStringSubmatch(text); len(match) > 2 {
		out.CompanyName = strings.TrimSpace(match[2])
	}

	return out
}

func (k KeyFields) MissingRequired() []string {
	missing := make([]string, 0, 3)
	if strings.TrimSpace(k.Amount) == "" {
		missing = append(missing, "amount")
	}
	if strings.TrimSpace(k.Date) == "" {
		missing = append(missing, "date")
	}
	if strings.TrimSpace(k.CompanyName) == "" {
		missing = append(missing, "company_name")
	}
	return missing
}

func (k KeyFields) FoundCount() int {
	count := 0
	if strings.TrimSpace(k.Amount) != "" {
		count++
	}
	if strings.TrimSpace(k.Date) != "" {
		count++
	}
	if strings.TrimSpace(k.CompanyName) != "" {
		count++
	}
	return count
}

func findDate(text string) string {
	if iso := isoDatePattern.FindString(text); iso != "" {
		return iso
	}
	raw := chileDatePattern.FindString(text)
	if raw == "" {
		return ""
	}
	parsed, err := time.Parse("02/01/2006", raw)
	if err != nil {
		return raw
	}
	return parsed.UTC().Format("2006-01-02")
}
