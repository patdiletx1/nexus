package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"nexus/api/internal/audit"
	"nexus/api/internal/idempotency"
	"nexus/api/internal/vault"
)

type fakeAuditService struct {
	lastQuery      audit.EventQuery
	lastCountQuery audit.EventQuery
	events         []audit.Event
	totalCount     int64
	countCalled    bool
	loggedEvents   []audit.Event
}

func (f *fakeAuditService) LogEvent(event audit.Event) {
	f.loggedEvents = append(f.loggedEvents, event)
}

func (f *fakeAuditService) ListEvents(query audit.EventQuery) []audit.Event {
	f.lastQuery = query
	if f.events != nil {
		return f.events
	}
	return []audit.Event{}
}

func (f *fakeAuditService) CountEvents(query audit.EventQuery) (int64, bool) {
	f.lastCountQuery = query
	f.countCalled = true
	return f.totalCount, true
}

type fakeIdempotencyService struct {
	records map[string]idempotency.StoredResponse
}

type fakeExtractor struct {
	documentType string
	extracted    string
	err          error
}

func (f fakeExtractor) ProcessItem(_ vault.Item, _ []byte) (string, string, error) {
	return f.documentType, f.extracted, f.err
}

func (f *fakeIdempotencyService) Get(_ context.Context, key idempotency.Key) (idempotency.StoredResponse, bool) {
	if f.records == nil {
		return idempotency.StoredResponse{}, false
	}
	response, ok := f.records[key.IdempotencyKey]
	return response, ok
}

func (f *fakeIdempotencyService) Put(_ context.Context, key idempotency.Key, response idempotency.StoredResponse, _ time.Duration) {
	if f.records == nil {
		f.records = map[string]idempotency.StoredResponse{}
	}
	f.records[key.IdempotencyKey] = response
}

func (f *fakeIdempotencyService) CleanupExpired(_ context.Context, _ int64) (int64, bool) {
	return 0, true
}

func TestVaultUploadAndList(t *testing.T) {
	handler := VaultHandler{Store: vault.NewInMemoryStore()}

	reqBody := []byte(`{
		"file_name":"presupuesto_base.pdf",
		"mime_type":"application/pdf",
		"size_bytes":1234,
		"sha256":"hash-123"
	}`)
	req := httptest.NewRequest(http.MethodPost, "/v1/vault/upload", bytes.NewReader(reqBody))
	req = req.WithContext(WithAuthContext(context.Background(), "user-1", "company-1", "member"))
	rec := httptest.NewRecorder()
	handler.Upload(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, rec.Code)
	}

	listReq := httptest.NewRequest(http.MethodGet, "/v1/vault/items", nil)
	listReq = listReq.WithContext(WithAuthContext(context.Background(), "user-1", "company-1", "member"))
	listRec := httptest.NewRecorder()
	handler.ListItems(listRec, listReq)

	if listRec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, listRec.Code)
	}
}

func TestVaultProcessLifecycle(t *testing.T) {
	store := vault.NewInMemoryStore()
	handler := VaultHandler{Store: store}

	item := vault.Item{
		ID:             "item-1",
		CompanyID:      "company-1",
		UploaderUserID: "user-1",
		StoragePath:    "company-1/item-1_presupuesto.pdf",
		FileName:       "presupuesto_obra.pdf",
		MimeType:       "application/pdf",
		SizeBytes:      5000,
		SHA256:         "sha-1",
		Status:         "uploaded",
		CreatedAt:      time.Now().UTC(),
		UpdatedAt:      time.Now().UTC(),
	}
	store.Save(item)

	processReqBody := []byte(`{"item_id":"item-1"}`)
	processReq := httptest.NewRequest(http.MethodPost, "/v1/vault/process", bytes.NewReader(processReqBody))
	processReq = processReq.WithContext(WithAuthContext(context.Background(), "user-1", "company-1", "member"))
	processRec := httptest.NewRecorder()
	handler.Process(processRec, processReq)

	if processRec.Code != http.StatusAccepted {
		t.Fatalf("expected status %d, got %d", http.StatusAccepted, processRec.Code)
	}

	time.Sleep(300 * time.Millisecond)
	processed, ok := store.GetByIDForCompany("item-1", "company-1")
	if !ok {
		t.Fatal("expected processed item to exist")
	}
	if processed.Status != "processed" {
		t.Fatalf("expected status processed, got %s", processed.Status)
	}
	if processed.DocumentType == "" || processed.ExtractedText == "" {
		t.Fatal("expected processing outputs to be populated")
	}
}

func TestVaultUploadRejectsInvalidMimeType(t *testing.T) {
	handler := VaultHandler{Store: vault.NewInMemoryStore()}

	reqBody := map[string]any{
		"file_name":  "script.sh",
		"mime_type":  "text/plain",
		"size_bytes": 99,
		"sha256":     "h",
	}
	payload, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/v1/vault/upload", bytes.NewReader(payload))
	req = req.WithContext(WithAuthContext(context.Background(), "user-1", "company-1", "member"))
	rec := httptest.NewRecorder()
	handler.Upload(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
}

func TestVaultGetItemByID(t *testing.T) {
	store := vault.NewInMemoryStore()
	now := time.Now().UTC()
	store.Save(vault.Item{
		ID:             "item-abc",
		CompanyID:      "company-1",
		UploaderUserID: "user-1",
		StoragePath:    "company-1/item-abc_factura.pdf",
		FileName:       "factura.pdf",
		MimeType:       "application/pdf",
		SizeBytes:      123,
		SHA256:         "sha-a",
		Status:         "uploaded",
		CreatedAt:      now,
		UpdatedAt:      now,
	})

	handler := VaultHandler{Store: store}
	req := httptest.NewRequest(http.MethodGet, "/v1/vault/items/item-abc", nil)
	req.SetPathValue("id", "item-abc")
	req = req.WithContext(WithAuthContext(context.Background(), "user-1", "company-1", "member"))
	rec := httptest.NewRecorder()
	handler.GetItem(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
}

func TestVaultGetItemEvents(t *testing.T) {
	store := vault.NewInMemoryStore()
	auditSvc := &fakeAuditService{
		events: []audit.Event{
			{
				ID:         "11111111-1111-1111-1111-111111111111",
				CompanyID:  "company-1",
				EntityType: "vault_item",
				EntityID:   "item-evt",
				EventType:  "vault_item_uploaded",
				CreatedAt:  time.Date(2026, 4, 10, 10, 0, 0, 0, time.UTC),
			},
		},
		totalCount: 37,
	}
	now := time.Now().UTC()
	store.Save(vault.Item{
		ID:             "item-evt",
		CompanyID:      "company-1",
		UploaderUserID: "user-1",
		StoragePath:    "company-1/item-evt_presupuesto.pdf",
		FileName:       "presupuesto.pdf",
		MimeType:       "application/pdf",
		SizeBytes:      456,
		SHA256:         "sha-b",
		Status:         "processed",
		CreatedAt:      now,
		UpdatedAt:      now,
	})

	handler := VaultHandler{Store: store, Audit: auditSvc}
	beforeCursor := encodeBeforeCursor(time.Date(2026, 4, 12, 0, 0, 0, 0, time.UTC), "11111111-1111-1111-1111-111111111111")
	req := httptest.NewRequest(
		http.MethodGet,
		"/v1/vault/items/item-evt/events?limit=10&event_type=vault_item_uploaded&from=2026-04-10T00:00:00Z&to=2026-04-11T00:00:00Z&before_cursor="+beforeCursor+"&include_total=true",
		nil,
	)
	req.SetPathValue("id", "item-evt")
	req = req.WithContext(WithAuthContext(context.Background(), "user-1", "company-1", "member"))
	rec := httptest.NewRecorder()
	handler.GetItemEvents(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
	if auditSvc.lastQuery.Limit != 10 {
		t.Fatalf("expected limit 10, got %d", auditSvc.lastQuery.Limit)
	}
	if auditSvc.lastQuery.EventType != "vault_item_uploaded" {
		t.Fatalf("expected event type filter, got %s", auditSvc.lastQuery.EventType)
	}
	if auditSvc.lastQuery.From == nil || auditSvc.lastQuery.To == nil {
		t.Fatal("expected range time filters to be parsed")
	}
	if auditSvc.lastQuery.BeforeCreatedAt == nil || auditSvc.lastQuery.BeforeEventID == "" {
		t.Fatal("expected composed cursor to be parsed")
	}
	if !auditSvc.countCalled {
		t.Fatal("expected CountEvents to be called when include_total=true")
	}
	if auditSvc.lastCountQuery.BeforeCreatedAt != nil || auditSvc.lastCountQuery.BeforeEventID != "" {
		t.Fatal("expected count query to ignore pagination cursor")
	}

	var payload map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("failed decoding payload: %v", err)
	}
	returnedCount, ok := payload["returned_count"].(float64)
	if !ok || int(returnedCount) != 1 {
		t.Fatalf("expected returned_count 1, got %v", payload["returned_count"])
	}
	hasMore, ok := payload["has_more"].(bool)
	if !ok || hasMore {
		t.Fatalf("expected has_more false, got %v", payload["has_more"])
	}
	totalCount, ok := payload["total_count"].(float64)
	if !ok || int(totalCount) != 37 {
		t.Fatalf("expected total_count 37, got %v", payload["total_count"])
	}
}

func TestVaultGetItemEventsInvalidLimit(t *testing.T) {
	store := vault.NewInMemoryStore()
	now := time.Now().UTC()
	store.Save(vault.Item{
		ID:             "item-evt-invalid",
		CompanyID:      "company-1",
		UploaderUserID: "user-1",
		StoragePath:    "company-1/item-evt-invalid.pdf",
		FileName:       "doc.pdf",
		MimeType:       "application/pdf",
		SizeBytes:      100,
		SHA256:         "sha-c",
		Status:         "uploaded",
		CreatedAt:      now,
		UpdatedAt:      now,
	})

	handler := VaultHandler{Store: store}
	req := httptest.NewRequest(http.MethodGet, "/v1/vault/items/item-evt-invalid/events?limit=abc", nil)
	req.SetPathValue("id", "item-evt-invalid")
	req = req.WithContext(WithAuthContext(context.Background(), "user-1", "company-1", "member"))
	rec := httptest.NewRecorder()
	handler.GetItemEvents(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
}

func TestVaultGetItemEventsInvalidTimeFilter(t *testing.T) {
	store := vault.NewInMemoryStore()
	now := time.Now().UTC()
	store.Save(vault.Item{
		ID:             "item-time-invalid",
		CompanyID:      "company-1",
		UploaderUserID: "user-1",
		StoragePath:    "company-1/item-time-invalid.pdf",
		FileName:       "doc.pdf",
		MimeType:       "application/pdf",
		SizeBytes:      100,
		SHA256:         "sha-d",
		Status:         "uploaded",
		CreatedAt:      now,
		UpdatedAt:      now,
	})

	handler := VaultHandler{Store: store}
	req := httptest.NewRequest(http.MethodGet, "/v1/vault/items/item-time-invalid/events?from=bad-date", nil)
	req.SetPathValue("id", "item-time-invalid")
	req = req.WithContext(WithAuthContext(context.Background(), "user-1", "company-1", "member"))
	rec := httptest.NewRecorder()
	handler.GetItemEvents(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
}

func TestVaultGetItemEventsInvalidCursor(t *testing.T) {
	store := vault.NewInMemoryStore()
	now := time.Now().UTC()
	store.Save(vault.Item{
		ID:             "item-cursor-invalid",
		CompanyID:      "company-1",
		UploaderUserID: "user-1",
		StoragePath:    "company-1/item-cursor-invalid.pdf",
		FileName:       "doc.pdf",
		MimeType:       "application/pdf",
		SizeBytes:      100,
		SHA256:         "sha-e",
		Status:         "uploaded",
		CreatedAt:      now,
		UpdatedAt:      now,
	})

	handler := VaultHandler{Store: store}
	req := httptest.NewRequest(http.MethodGet, "/v1/vault/items/item-cursor-invalid/events?before_cursor=bad", nil)
	req.SetPathValue("id", "item-cursor-invalid")
	req = req.WithContext(WithAuthContext(context.Background(), "user-1", "company-1", "member"))
	rec := httptest.NewRecorder()
	handler.GetItemEvents(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
}

func TestVaultRetryItemFromFailed(t *testing.T) {
	store := vault.NewInMemoryStore()
	now := time.Now().UTC()
	store.Save(vault.Item{
		ID:             "item-retry-ok",
		CompanyID:      "company-1",
		UploaderUserID: "user-1",
		StoragePath:    "company-1/item-retry-ok.pdf",
		FileName:       "presupuesto_retry.pdf",
		MimeType:       "application/pdf",
		SizeBytes:      123,
		SHA256:         "sha-f",
		Status:         "failed",
		CreatedAt:      now,
		UpdatedAt:      now,
	})

	handler := VaultHandler{Store: store}
	req := httptest.NewRequest(http.MethodPost, "/v1/vault/items/item-retry-ok/retry", nil)
	req.SetPathValue("id", "item-retry-ok")
	req = req.WithContext(WithAuthContext(context.Background(), "user-1", "company-1", "member"))
	rec := httptest.NewRecorder()
	handler.RetryItem(rec, req)

	if rec.Code != http.StatusAccepted {
		t.Fatalf("expected status %d, got %d", http.StatusAccepted, rec.Code)
	}
}

func TestVaultRetryItemRejectsNonFailedStatus(t *testing.T) {
	store := vault.NewInMemoryStore()
	now := time.Now().UTC()
	store.Save(vault.Item{
		ID:             "item-retry-bad",
		CompanyID:      "company-1",
		UploaderUserID: "user-1",
		StoragePath:    "company-1/item-retry-bad.pdf",
		FileName:       "presupuesto_retry_bad.pdf",
		MimeType:       "application/pdf",
		SizeBytes:      123,
		SHA256:         "sha-g",
		Status:         "processed",
		CreatedAt:      now,
		UpdatedAt:      now,
	})

	handler := VaultHandler{Store: store}
	req := httptest.NewRequest(http.MethodPost, "/v1/vault/items/item-retry-bad/retry", nil)
	req.SetPathValue("id", "item-retry-bad")
	req = req.WithContext(WithAuthContext(context.Background(), "user-1", "company-1", "member"))
	rec := httptest.NewRecorder()
	handler.RetryItem(rec, req)

	if rec.Code != http.StatusConflict {
		t.Fatalf("expected status %d, got %d", http.StatusConflict, rec.Code)
	}
}

func TestVaultProcessRejectsAlreadyProcessing(t *testing.T) {
	store := vault.NewInMemoryStore()
	now := time.Now().UTC()
	store.Save(vault.Item{
		ID:             "item-processing",
		CompanyID:      "company-1",
		UploaderUserID: "user-1",
		StoragePath:    "company-1/item-processing.pdf",
		FileName:       "doc.pdf",
		MimeType:       "application/pdf",
		SizeBytes:      100,
		SHA256:         "sha-h",
		Status:         "processing",
		CreatedAt:      now,
		UpdatedAt:      now,
	})

	handler := VaultHandler{Store: store}
	req := httptest.NewRequest(http.MethodPost, "/v1/vault/process", bytes.NewReader([]byte(`{"item_id":"item-processing"}`)))
	req = req.WithContext(WithAuthContext(context.Background(), "user-1", "company-1", "member"))
	rec := httptest.NewRecorder()
	handler.Process(rec, req)

	if rec.Code != http.StatusConflict {
		t.Fatalf("expected status %d, got %d", http.StatusConflict, rec.Code)
	}
}

func TestVaultProcessRejectsNonUploadedStatus(t *testing.T) {
	store := vault.NewInMemoryStore()
	now := time.Now().UTC()
	store.Save(vault.Item{
		ID:             "item-processed",
		CompanyID:      "company-1",
		UploaderUserID: "user-1",
		StoragePath:    "company-1/item-processed.pdf",
		FileName:       "doc.pdf",
		MimeType:       "application/pdf",
		SizeBytes:      100,
		SHA256:         "sha-i",
		Status:         "processed",
		CreatedAt:      now,
		UpdatedAt:      now,
	})

	handler := VaultHandler{Store: store}
	req := httptest.NewRequest(http.MethodPost, "/v1/vault/process", bytes.NewReader([]byte(`{"item_id":"item-processed"}`)))
	req = req.WithContext(WithAuthContext(context.Background(), "user-1", "company-1", "member"))
	rec := httptest.NewRecorder()
	handler.Process(rec, req)

	if rec.Code != http.StatusConflict {
		t.Fatalf("expected status %d, got %d", http.StatusConflict, rec.Code)
	}
}

func TestVaultProcessIdempotencyReplay(t *testing.T) {
	store := vault.NewInMemoryStore()
	idem := &fakeIdempotencyService{}
	now := time.Now().UTC()
	store.Save(vault.Item{
		ID:             "item-idem-process",
		CompanyID:      "company-1",
		UploaderUserID: "user-1",
		StoragePath:    "company-1/item-idem-process.pdf",
		FileName:       "presupuesto_idem.pdf",
		MimeType:       "application/pdf",
		SizeBytes:      100,
		SHA256:         "sha-idem",
		Status:         "uploaded",
		CreatedAt:      now,
		UpdatedAt:      now,
	})

	handler := VaultHandler{Store: store, Idempotency: idem}
	req1 := httptest.NewRequest(http.MethodPost, "/v1/vault/process", bytes.NewReader([]byte(`{"item_id":"item-idem-process"}`)))
	req1.Header.Set("Idempotency-Key", "idem-process-1")
	req1 = req1.WithContext(WithAuthContext(context.Background(), "user-1", "company-1", "member"))
	rec1 := httptest.NewRecorder()
	handler.Process(rec1, req1)
	if rec1.Code != http.StatusAccepted {
		t.Fatalf("expected first status %d, got %d", http.StatusAccepted, rec1.Code)
	}

	req2 := httptest.NewRequest(http.MethodPost, "/v1/vault/process", bytes.NewReader([]byte(`{"item_id":"item-idem-process"}`)))
	req2.Header.Set("Idempotency-Key", "idem-process-1")
	req2 = req2.WithContext(WithAuthContext(context.Background(), "user-1", "company-1", "member"))
	rec2 := httptest.NewRecorder()
	handler.Process(rec2, req2)
	if rec2.Code != http.StatusAccepted {
		t.Fatalf("expected replay status %d, got %d", http.StatusAccepted, rec2.Code)
	}
}

func TestVaultRetryIdempotencyReplay(t *testing.T) {
	store := vault.NewInMemoryStore()
	idem := &fakeIdempotencyService{}
	now := time.Now().UTC()
	store.Save(vault.Item{
		ID:             "item-idem-retry",
		CompanyID:      "company-1",
		UploaderUserID: "user-1",
		StoragePath:    "company-1/item-idem-retry.pdf",
		FileName:       "retry.pdf",
		MimeType:       "application/pdf",
		SizeBytes:      100,
		SHA256:         "sha-idem-retry",
		Status:         "failed",
		CreatedAt:      now,
		UpdatedAt:      now,
	})

	handler := VaultHandler{Store: store, Idempotency: idem}
	req1 := httptest.NewRequest(http.MethodPost, "/v1/vault/items/item-idem-retry/retry", nil)
	req1.SetPathValue("id", "item-idem-retry")
	req1.Header.Set("Idempotency-Key", "idem-retry-1")
	req1 = req1.WithContext(WithAuthContext(context.Background(), "user-1", "company-1", "member"))
	rec1 := httptest.NewRecorder()
	handler.RetryItem(rec1, req1)
	if rec1.Code != http.StatusAccepted {
		t.Fatalf("expected first status %d, got %d", http.StatusAccepted, rec1.Code)
	}

	req2 := httptest.NewRequest(http.MethodPost, "/v1/vault/items/item-idem-retry/retry", nil)
	req2.SetPathValue("id", "item-idem-retry")
	req2.Header.Set("Idempotency-Key", "idem-retry-1")
	req2 = req2.WithContext(WithAuthContext(context.Background(), "user-1", "company-1", "member"))
	rec2 := httptest.NewRecorder()
	handler.RetryItem(rec2, req2)
	if rec2.Code != http.StatusAccepted {
		t.Fatalf("expected replay status %d, got %d", http.StatusAccepted, rec2.Code)
	}
}

func TestVaultProcessFailureAddsErrorClassification(t *testing.T) {
	store := vault.NewInMemoryStore()
	auditSvc := &fakeAuditService{}
	now := time.Now().UTC()
	store.Save(vault.Item{
		ID:             "item-error-category",
		CompanyID:      "company-1",
		UploaderUserID: "user-1",
		StoragePath:    "company-1/item-error-category.bin",
		FileName:       "adjunto.bin",
		MimeType:       "application/octet-stream",
		SizeBytes:      100,
		SHA256:         "sha-error-category",
		Status:         "uploaded",
		CreatedAt:      now,
		UpdatedAt:      now,
	})

	handler := VaultHandler{Store: store, Audit: auditSvc}
	req := httptest.NewRequest(http.MethodPost, "/v1/vault/process", bytes.NewReader([]byte(`{"item_id":"item-error-category"}`)))
	req = req.WithContext(WithAuthContext(context.Background(), "user-1", "company-1", "member"))
	rec := httptest.NewRecorder()
	handler.Process(rec, req)
	if rec.Code != http.StatusAccepted {
		t.Fatalf("expected first status %d, got %d", http.StatusAccepted, rec.Code)
	}

	time.Sleep(350 * time.Millisecond)

	processed, ok := store.GetByIDForCompany("item-error-category", "company-1")
	if !ok {
		t.Fatal("expected item to exist")
	}
	if processed.Status != "failed" {
		t.Fatalf("expected failed status, got %s", processed.Status)
	}

	var failurePayload map[string]any
	for _, event := range auditSvc.loggedEvents {
		if event.EventType == "vault_item_processing_failed" {
			failurePayload = event.Payload
		}
	}
	if failurePayload == nil {
		t.Fatal("expected vault_item_processing_failed event to be logged")
	}
	if failurePayload["error_category"] != "unsupported_document" {
		t.Fatalf("expected unsupported_document category, got %v", failurePayload["error_category"])
	}
	if failurePayload["error_stage"] != "extract_content" {
		t.Fatalf("expected extract_content stage, got %v", failurePayload["error_stage"])
	}
	if failurePayload["document_family"] != "other" {
		t.Fatalf("expected document family other, got %v", failurePayload["document_family"])
	}
	retryable, ok := failurePayload["retryable"].(bool)
	if !ok {
		t.Fatalf("expected retryable bool in payload, got %T", failurePayload["retryable"])
	}
	if retryable {
		t.Fatal("expected unsupported document error to be non-retryable")
	}
}

func TestVaultProcessAddsKeyFieldsToProcessedEvent(t *testing.T) {
	store := vault.NewInMemoryStore()
	auditSvc := &fakeAuditService{}
	now := time.Now().UTC()
	store.Save(vault.Item{
		ID:             "item-key-fields",
		CompanyID:      "company-1",
		UploaderUserID: "user-1",
		StoragePath:    "company-1/item-key-fields.pdf",
		FileName:       "presupuesto_key_fields.pdf",
		MimeType:       "application/pdf",
		SizeBytes:      100,
		SHA256:         "sha-key-fields",
		Status:         "uploaded",
		CreatedAt:      now,
		UpdatedAt:      now,
	})

	handler := VaultHandler{
		Store: store,
		Audit: auditSvc,
		Extractor: fakeExtractor{
			documentType: "presupuesto",
			extracted:    "Razon social: Acme Limitada\nMonto total: $4.200.000,00\nFecha cierre: 2026-04-20",
		},
	}
	req := httptest.NewRequest(http.MethodPost, "/v1/vault/process", bytes.NewReader([]byte(`{"item_id":"item-key-fields"}`)))
	req = req.WithContext(WithAuthContext(context.Background(), "user-1", "company-1", "member"))
	rec := httptest.NewRecorder()
	handler.Process(rec, req)
	if rec.Code != http.StatusAccepted {
		t.Fatalf("expected first status %d, got %d", http.StatusAccepted, rec.Code)
	}

	time.Sleep(100 * time.Millisecond)

	var processedPayload map[string]any
	for _, event := range auditSvc.loggedEvents {
		if event.EventType == "vault_item_processed" {
			processedPayload = event.Payload
		}
	}
	if processedPayload == nil {
		t.Fatal("expected vault_item_processed event to be logged")
	}
	if processedPayload["key_fields_found"] != 3 {
		t.Fatalf("expected key_fields_found=3, got %v", processedPayload["key_fields_found"])
	}

	keyFields, ok := processedPayload["key_fields"].(map[string]any)
	if !ok {
		t.Fatalf("expected key_fields map, got %T", processedPayload["key_fields"])
	}
	if keyFields["company_name"] != "Acme Limitada" {
		t.Fatalf("unexpected company_name: %v", keyFields["company_name"])
	}
	if keyFields["amount"] != "$4.200.000,00" {
		t.Fatalf("unexpected amount: %v", keyFields["amount"])
	}
	if keyFields["date"] != "2026-04-20" {
		t.Fatalf("unexpected date: %v", keyFields["date"])
	}

	missingFields, ok := processedPayload["missing_key_fields"].([]string)
	if !ok {
		t.Fatalf("expected []string missing_key_fields, got %T", processedPayload["missing_key_fields"])
	}
	if len(missingFields) != 0 {
		t.Fatalf("expected no missing fields, got %v", missingFields)
	}
}
