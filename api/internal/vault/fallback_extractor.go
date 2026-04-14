package vault

import (
	"fmt"
	"strings"
)

type NamedExtractor struct {
	Name      string
	Extractor Extractor
}

type FallbackStrategy struct {
	Primary         NamedExtractor
	Secondary       NamedExtractor
	RequireManualOn bool
}

type FallbackMatrixExtractor struct {
	ByDocumentFamily map[string]FallbackStrategy
	DefaultStrategy  FallbackStrategy
}

func NewFallbackMatrixExtractor(byFamily map[string]FallbackStrategy, defaultStrategy FallbackStrategy) FallbackMatrixExtractor {
	return FallbackMatrixExtractor{
		ByDocumentFamily: byFamily,
		DefaultStrategy:  defaultStrategy,
	}
}

func (f FallbackMatrixExtractor) ProcessItem(item Item, content []byte) (string, string, error) {
	strategy := f.strategyForFamily(DocumentFamily(item.MimeType))

	if strategy.Primary.Extractor != nil {
		docType, extractedText, err := strategy.Primary.Extractor.ProcessItem(item, content)
		if err == nil {
			return docType, extractedText, nil
		}

		primaryErr := err
		if strategy.Secondary.Extractor != nil {
			docType, extractedText, secondaryErr := strategy.Secondary.Extractor.ProcessItem(item, content)
			if secondaryErr == nil {
				text := extractedText
				if strings.TrimSpace(text) != "" {
					text = fmt.Sprintf("[fallback_provider=%s] %s", strategy.Secondary.Name, text)
				}
				return docType, text, nil
			}
			if strategy.RequireManualOn {
				return "", "", fmt.Errorf(
					"manual_review_required family=%s primary=%s primary_error=%v secondary=%s secondary_error=%v",
					DocumentFamily(item.MimeType),
					strategy.Primary.Name,
					primaryErr,
					strategy.Secondary.Name,
					secondaryErr,
				)
			}

			return "", "", fmt.Errorf(
				"processing_failed family=%s primary=%s primary_error=%v secondary=%s secondary_error=%v",
				DocumentFamily(item.MimeType),
				strategy.Primary.Name,
				primaryErr,
				strategy.Secondary.Name,
				secondaryErr,
			)
		}

		if strategy.RequireManualOn {
			return "", "", fmt.Errorf(
				"manual_review_required family=%s primary=%s primary_error=%v",
				DocumentFamily(item.MimeType),
				strategy.Primary.Name,
				primaryErr,
			)
		}
		return "", "", primaryErr
	}

	if strategy.RequireManualOn {
		return "", "", fmt.Errorf(
			"manual_review_required family=%s reason=no_provider_configured",
			DocumentFamily(item.MimeType),
		)
	}

	return "", "", fmt.Errorf("no_extractor_available_for_family=%s", DocumentFamily(item.MimeType))
}

func (f FallbackMatrixExtractor) strategyForFamily(family string) FallbackStrategy {
	if strategy, ok := f.ByDocumentFamily[family]; ok {
		return strategy
	}
	return f.DefaultStrategy
}
