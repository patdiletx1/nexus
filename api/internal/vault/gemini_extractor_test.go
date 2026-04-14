package vault

import "testing"

func TestSupportsInlineData(t *testing.T) {
	if !supportsInlineData("application/pdf") {
		t.Fatal("expected pdf to support inline data")
	}
	if !supportsInlineData("image/png") {
		t.Fatal("expected image to support inline data")
	}
	if !supportsInlineData("audio/mpeg") {
		t.Fatal("expected audio to support inline data")
	}
	if supportsInlineData("text/plain") {
		t.Fatal("expected text/plain to not support inline data")
	}
}

func TestContentPreviewForPrompt(t *testing.T) {
	textPreview := contentPreviewForPrompt("text/plain", []byte("hola mundo"))
	if textPreview != "hola mundo" {
		t.Fatalf("unexpected text preview: %s", textPreview)
	}

	binaryPreview := contentPreviewForPrompt("application/pdf", []byte{1, 2, 3, 4})
	if binaryPreview == "" {
		t.Fatal("expected binary preview description")
	}
}
