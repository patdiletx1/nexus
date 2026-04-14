package vault

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

const maxInlineDataBytes = 2 * 1024 * 1024

type GeminiExtractor struct {
	APIKey     string
	Model      string
	HTTPClient *http.Client
}

func NewGeminiExtractor(apiKey, model string) GeminiExtractor {
	if model == "" {
		model = "gemini-1.5-flash"
	}
	return GeminiExtractor{
		APIKey: apiKey,
		Model:  model,
	}
}

func (g GeminiExtractor) ProcessItem(item Item, content []byte) (string, string, error) {
	client := g.HTTPClient
	if client == nil {
		client = &http.Client{Timeout: 15 * time.Second}
	}

	prompt := fmt.Sprintf(
		"Clasifica el tipo documental probable y resume texto extraido para este archivo. Responde JSON con claves doc_type y extracted_text. file_name=%s mime_type=%s sha256=%s content_size_bytes=%d preview=%q",
		item.FileName,
		item.MimeType,
		item.SHA256,
		len(content),
		contentPreviewForPrompt(item.MimeType, content),
	)

	parts := []map[string]any{
		{"text": prompt},
	}
	if supportsInlineData(item.MimeType) && len(content) > 0 {
		inlineBytes := content
		if len(inlineBytes) > maxInlineDataBytes {
			inlineBytes = inlineBytes[:maxInlineDataBytes]
		}

		parts = append(parts, map[string]any{
			"inline_data": map[string]string{
				"mime_type": item.MimeType,
				"data":      base64.StdEncoding.EncodeToString(inlineBytes),
			},
		})
	}

	reqBody := map[string]any{
		"contents": []map[string]any{
			{
				"parts": parts,
			},
		},
	}
	payload, _ := json.Marshal(reqBody)

	url := fmt.Sprintf(
		"https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent?key=%s",
		g.Model,
		g.APIKey,
	)
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(payload))
	if err != nil {
		return "", "", err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return "", "", fmt.Errorf("gemini request failed with status %d", resp.StatusCode)
	}

	var out struct {
		Candidates []struct {
			Content struct {
				Parts []struct {
					Text string `json:"text"`
				} `json:"parts"`
			} `json:"content"`
		} `json:"candidates"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return "", "", err
	}
	if len(out.Candidates) == 0 || len(out.Candidates[0].Content.Parts) == 0 {
		return "", "", fmt.Errorf("gemini response empty")
	}

	text := strings.TrimSpace(out.Candidates[0].Content.Parts[0].Text)
	docType := detectDocumentType(item.FileName, item.MimeType)

	var parsed struct {
		DocType       string `json:"doc_type"`
		ExtractedText string `json:"extracted_text"`
	}
	if err := json.Unmarshal([]byte(stripCodeFence(text)), &parsed); err == nil {
		if strings.TrimSpace(parsed.DocType) != "" {
			docType = strings.TrimSpace(parsed.DocType)
		}
		if strings.TrimSpace(parsed.ExtractedText) != "" {
			return docType, strings.TrimSpace(parsed.ExtractedText), nil
		}
	}

	return docType, text, nil
}

func stripCodeFence(value string) string {
	value = strings.TrimSpace(value)
	value = strings.TrimPrefix(value, "```json")
	value = strings.TrimPrefix(value, "```")
	value = strings.TrimSuffix(value, "```")
	return strings.TrimSpace(value)
}

func truncateString(value string, maxLen int) string {
	if len(value) <= maxLen {
		return value
	}
	return value[:maxLen]
}

func supportsInlineData(mimeType string) bool {
	return strings.HasPrefix(mimeType, "image/") ||
		strings.HasPrefix(mimeType, "audio/") ||
		mimeType == "application/pdf"
}

func contentPreviewForPrompt(mimeType string, content []byte) string {
	if len(content) == 0 {
		return ""
	}
	if isLikelyTextMimeType(mimeType) {
		return truncateString(string(content), 1200)
	}

	return fmt.Sprintf("binary_content_attached=%t bytes=%d", supportsInlineData(mimeType), len(content))
}

func isLikelyTextMimeType(mimeType string) bool {
	return strings.HasPrefix(mimeType, "text/") ||
		mimeType == "application/json" ||
		mimeType == "application/xml" ||
		mimeType == "text/csv"
}
