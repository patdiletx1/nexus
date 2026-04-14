package vault

import (
	"errors"
	"strings"
)

type ProcessingErrorDetails struct {
	Stage     string
	Category  string
	Retryable bool
}

func ClassifyProcessingError(stage string, err error) ProcessingErrorDetails {
	message := strings.ToLower(strings.TrimSpace(errorMessage(err)))
	details := ProcessingErrorDetails{
		Stage:     stage,
		Category:  "unknown",
		Retryable: true,
	}

	switch {
	case containsAny(message, "manual_review_required"):
		details.Category = "manual_review_required"
		details.Retryable = false
	case containsAny(message, "timeout", "deadline exceeded", "timed out"):
		details.Category = "timeout"
		details.Retryable = true
	case containsAny(message, "connection reset", "connection refused", "temporarily unavailable", "network", "unexpected eof"):
		details.Category = "transient_upstream"
		details.Retryable = true
	case containsAny(message, "not configured", "missing", "forbidden", "unauthorized", "permission denied", "invalid api key"):
		details.Category = "configuration"
		details.Retryable = false
	case containsAny(message, "too large", "size limit", "exceeds"):
		details.Category = "payload_too_large"
		details.Retryable = false
	case containsAny(message, "unsupported", "unable to classify", "invalid mime", "mime type not allowed"):
		details.Category = "unsupported_document"
		details.Retryable = false
	}

	return details
}

func DocumentFamily(mimeType string) string {
	switch {
	case mimeType == "application/pdf":
		return "pdf"
	case strings.HasPrefix(mimeType, "image/"):
		return "image"
	case strings.HasPrefix(mimeType, "audio/"):
		return "audio"
	default:
		return "other"
	}
}

func errorMessage(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}

func containsAny(value string, fragments ...string) bool {
	for _, fragment := range fragments {
		if strings.Contains(value, fragment) {
			return true
		}
	}
	return false
}

func IsRetryableProcessingError(stage string, err error) bool {
	return ClassifyProcessingError(stage, err).Retryable
}

func IsUnsupportedDocumentError(err error) bool {
	return ClassifyProcessingError("extract_content", err).Category == "unsupported_document"
}

var ErrUnsupportedDocument = errors.New("unsupported document")
