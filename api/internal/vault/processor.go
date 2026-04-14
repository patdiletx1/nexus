package vault

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

type Extractor interface {
	ProcessItem(item Item, content []byte) (string, string, error)
}

type SimulatedExtractor struct{}

func NewSimulatedExtractor() SimulatedExtractor {
	return SimulatedExtractor{}
}

func (SimulatedExtractor) ProcessItem(item Item, content []byte) (string, string, error) {
	time.Sleep(250 * time.Millisecond)

	documentType := detectDocumentType(item.FileName, item.MimeType)
	if documentType == "otro" {
		return "", "", errors.New("unable to classify document type")
	}

	extractedText := fmt.Sprintf("Simulated extraction for file=%s mime=%s sha256=%s", item.FileName, item.MimeType, item.SHA256)
	if len(content) > 0 {
		extractedText = fmt.Sprintf("%s content_bytes=%d", extractedText, len(content))
	}

	return documentType, extractedText, nil
}

func detectDocumentType(fileName, mimeType string) string {
	name := strings.ToLower(fileName)

	switch {
	case strings.Contains(name, "presupuesto"):
		return "presupuesto"
	case strings.Contains(name, "certificado"):
		return "certificado"
	case strings.Contains(name, "plano"):
		return "plano"
	case strings.Contains(name, "factura"):
		return "factura"
	case strings.HasPrefix(mimeType, "audio/"):
		return "audio_nota"
	case strings.HasPrefix(mimeType, "image/"):
		return "documento_imagen"
	case mimeType == "application/pdf":
		return "documento_pdf"
	default:
		return "otro"
	}
}
