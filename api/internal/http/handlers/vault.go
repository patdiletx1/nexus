package handlers

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"time"

	"nexus/api/internal/audit"
	"nexus/api/internal/idempotency"
	"nexus/api/internal/observability"
	"nexus/api/internal/storage"
	"nexus/api/internal/vault"
)

const maxUploadSizeBytes int64 = 25 * 1024 * 1024

var allowedMimeTypes = map[string]struct{}{
	"application/pdf": {},
	"image/jpeg":      {},
	"image/png":       {},
	"audio/mpeg":      {},
	"audio/mp4":       {},
	"audio/wav":       {},
}

type VaultHandler struct {
	Store       vault.Store
	Signer      storage.Signer
	Reader      storage.ObjectReader
	Extractor   vault.Extractor
	Audit       audit.Service
	Idempotency idempotency.Service
	Metrics     *observability.Metrics
}

type uploadRequest struct {
	FileName  string `json:"file_name"`
	MimeType  string `json:"mime_type"`
	SizeBytes int64  `json:"size_bytes"`
	SHA256    string `json:"sha256"`
}

type uploadResponse struct {
	ItemID      string    `json:"item_id"`
	StoragePath string    `json:"storage_path"`
	Status      string    `json:"status"`
	UploadURL   string    `json:"upload_url"`
	ExpiresAt   time.Time `json:"expires_at"`
}

type processRequest struct {
	ItemID string `json:"item_id"`
}

type processResponse struct {
	ItemID    string    `json:"item_id"`
	Status    string    `json:"status"`
	StartedAt time.Time `json:"started_at"`
}

func (h VaultHandler) Upload(w http.ResponseWriter, r *http.Request) {
	h = h.withDefaults()
	if h.Store == nil {
		http.Error(w, "vault store unavailable", http.StatusInternalServerError)
		return
	}

	var req uploadRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	req.FileName = strings.TrimSpace(req.FileName)
	req.MimeType = strings.TrimSpace(req.MimeType)
	req.SHA256 = strings.TrimSpace(req.SHA256)

	if req.FileName == "" || req.MimeType == "" || req.SHA256 == "" || req.SizeBytes <= 0 {
		http.Error(w, "missing required fields", http.StatusBadRequest)
		return
	}
	if req.SizeBytes > maxUploadSizeBytes {
		http.Error(w, "file exceeds maximum allowed size", http.StatusBadRequest)
		return
	}
	if _, ok := allowedMimeTypes[req.MimeType]; !ok {
		http.Error(w, "mime type not allowed", http.StatusBadRequest)
		return
	}

	userID := UserIDFromContext(r.Context())
	companyID := CompanyIDFromContext(r.Context())
	if userID == "" || companyID == "" {
		http.Error(w, "missing auth context", http.StatusUnauthorized)
		return
	}

	itemID := newRandomID()
	storagePath := fmt.Sprintf("%s/%s_%s", companyID, itemID, sanitizeFileName(req.FileName))
	now := time.Now().UTC()
	signed, err := h.Signer.SignUpload(storagePath, 900)
	if err != nil {
		http.Error(w, "failed to create upload url", http.StatusInternalServerError)
		return
	}

	h.Store.Save(vault.Item{
		ID:             itemID,
		CompanyID:      companyID,
		UploaderUserID: userID,
		StoragePath:    storagePath,
		FileName:       req.FileName,
		MimeType:       req.MimeType,
		SizeBytes:      req.SizeBytes,
		SHA256:         req.SHA256,
		Status:         "uploaded",
		CreatedAt:      now,
	})
	h.Audit.LogEvent(audit.Event{
		CompanyID:   companyID,
		ActorUserID: userID,
		EventType:   "vault_item_uploaded",
		EntityType:  "vault_item",
		EntityID:    itemID,
		Payload: map[string]any{
			"status":       "uploaded",
			"mime_type":    req.MimeType,
			"size_bytes":   req.SizeBytes,
			"storage_path": storagePath,
		},
	})

	writeJSON(w, http.StatusCreated, uploadResponse{
		ItemID:      itemID,
		StoragePath: storagePath,
		Status:      "uploaded",
		UploadURL:   signed.URL,
		ExpiresAt:   signed.ExpiresAt,
	})
}

func (h VaultHandler) Process(w http.ResponseWriter, r *http.Request) {
	h = h.withDefaults()
	if h.Store == nil {
		http.Error(w, "vault store unavailable", http.StatusInternalServerError)
		return
	}

	companyID := CompanyIDFromContext(r.Context())
	if companyID == "" {
		http.Error(w, "missing auth context", http.StatusUnauthorized)
		return
	}

	var req processRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	req.ItemID = strings.TrimSpace(req.ItemID)
	if req.ItemID == "" {
		http.Error(w, "item_id is required", http.StatusBadRequest)
		return
	}
	userID := UserIDFromContext(r.Context())
	idemKey, hasIdempotencyKey := h.buildIdempotencyKey(r, companyID, userID, "vault.process", req.ItemID)
	if hasIdempotencyKey {
		if storedResponse, found := h.Idempotency.Get(r.Context(), idemKey); found {
			writeJSON(w, storedResponse.StatusCode, storedResponse.Payload)
			return
		}
	}
	currentItem, ok := h.Store.GetByIDForCompany(req.ItemID, companyID)
	if !ok {
		http.Error(w, "item not found", http.StatusNotFound)
		return
	}
	if currentItem.Status == "processing" {
		http.Error(w, "item is already processing", http.StatusConflict)
		return
	}
	if currentItem.Status != "uploaded" {
		http.Error(w, "only uploaded items can be processed", http.StatusConflict)
		return
	}

	item, startedAt, err := h.startProcessing(req.ItemID, companyID, userID, "vault_item_processing_started", []string{"uploaded"})
	if err == errItemNotFound {
		http.Error(w, "item not found", http.StatusNotFound)
		return
	}
	if err == errItemStatusConflict {
		http.Error(w, "item cannot transition to processing from current status", http.StatusConflict)
		return
	}
	if err != nil {
		http.Error(w, "failed to process item", http.StatusInternalServerError)
		return
	}

	response := processResponse{
		ItemID:    item.ID,
		Status:    "processing",
		StartedAt: startedAt,
	}
	if hasIdempotencyKey {
		h.Idempotency.Put(r.Context(), idemKey, idempotency.StoredResponse{
			StatusCode: http.StatusAccepted,
			Payload: map[string]any{
				"item_id":    response.ItemID,
				"status":     response.Status,
				"started_at": response.StartedAt,
			},
		}, 24*time.Hour)
	}

	writeJSON(w, http.StatusAccepted, response)
}

func (h VaultHandler) RetryItem(w http.ResponseWriter, r *http.Request) {
	h = h.withDefaults()
	if h.Store == nil {
		http.Error(w, "vault store unavailable", http.StatusInternalServerError)
		return
	}

	companyID := CompanyIDFromContext(r.Context())
	if companyID == "" {
		http.Error(w, "missing auth context", http.StatusUnauthorized)
		return
	}

	itemID := strings.TrimSpace(r.PathValue("id"))
	if itemID == "" {
		http.Error(w, "item id is required", http.StatusBadRequest)
		return
	}
	item, ok := h.Store.GetByIDForCompany(itemID, companyID)
	if !ok {
		http.Error(w, "item not found", http.StatusNotFound)
		return
	}

	userID := UserIDFromContext(r.Context())
	idemKey, hasIdempotencyKey := h.buildIdempotencyKey(r, companyID, userID, "vault.retry", itemID)
	if hasIdempotencyKey {
		if storedResponse, found := h.Idempotency.Get(r.Context(), idemKey); found {
			writeJSON(w, storedResponse.StatusCode, storedResponse.Payload)
			return
		}
	}
	if item.Status != "failed" {
		http.Error(w, "only failed items can be retried", http.StatusConflict)
		return
	}

	processingItem, startedAt, err := h.startProcessing(itemID, companyID, userID, "vault_item_processing_retried", []string{"failed"})
	if err == errItemNotFound {
		http.Error(w, "item not found", http.StatusNotFound)
		return
	}
	if err == errItemStatusConflict {
		http.Error(w, "item cannot transition to processing from current status", http.StatusConflict)
		return
	}
	if err != nil {
		http.Error(w, "failed to retry item", http.StatusInternalServerError)
		return
	}

	response := processResponse{
		ItemID:    processingItem.ID,
		Status:    "processing",
		StartedAt: startedAt,
	}
	if hasIdempotencyKey {
		h.Idempotency.Put(r.Context(), idemKey, idempotency.StoredResponse{
			StatusCode: http.StatusAccepted,
			Payload: map[string]any{
				"item_id":    response.ItemID,
				"status":     response.Status,
				"started_at": response.StartedAt,
			},
		}, 24*time.Hour)
	}

	writeJSON(w, http.StatusAccepted, response)
}

func (h VaultHandler) ListItems(w http.ResponseWriter, r *http.Request) {
	h = h.withDefaults()
	if h.Store == nil {
		http.Error(w, "vault store unavailable", http.StatusInternalServerError)
		return
	}

	companyID := CompanyIDFromContext(r.Context())
	if companyID == "" {
		http.Error(w, "missing auth context", http.StatusUnauthorized)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"items": h.Store.ListByCompany(companyID),
	})
}

func (h VaultHandler) GetItem(w http.ResponseWriter, r *http.Request) {
	h = h.withDefaults()
	if h.Store == nil {
		http.Error(w, "vault store unavailable", http.StatusInternalServerError)
		return
	}

	companyID := CompanyIDFromContext(r.Context())
	if companyID == "" {
		http.Error(w, "missing auth context", http.StatusUnauthorized)
		return
	}

	itemID := strings.TrimSpace(r.PathValue("id"))
	if itemID == "" {
		http.Error(w, "item id is required", http.StatusBadRequest)
		return
	}

	item, ok := h.Store.GetByIDForCompany(itemID, companyID)
	if !ok {
		http.Error(w, "item not found", http.StatusNotFound)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"item": item})
}

func (h VaultHandler) GetItemEvents(w http.ResponseWriter, r *http.Request) {
	h = h.withDefaults()
	if h.Store == nil {
		http.Error(w, "vault store unavailable", http.StatusInternalServerError)
		return
	}

	companyID := CompanyIDFromContext(r.Context())
	if companyID == "" {
		http.Error(w, "missing auth context", http.StatusUnauthorized)
		return
	}

	itemID := strings.TrimSpace(r.PathValue("id"))
	if itemID == "" {
		http.Error(w, "item id is required", http.StatusBadRequest)
		return
	}
	if _, ok := h.Store.GetByIDForCompany(itemID, companyID); !ok {
		http.Error(w, "item not found", http.StatusNotFound)
		return
	}

	limit := 50
	if rawLimit := strings.TrimSpace(r.URL.Query().Get("limit")); rawLimit != "" {
		parsedLimit, err := strconv.Atoi(rawLimit)
		if err != nil || parsedLimit <= 0 {
			http.Error(w, "invalid limit value", http.StatusBadRequest)
			return
		}
		limit = parsedLimit
	}

	var from *time.Time
	var to *time.Time
	var err error
	if from, err = parseOptionalTime(r.URL.Query().Get("from")); err != nil {
		http.Error(w, "invalid from value, expected RFC3339", http.StatusBadRequest)
		return
	}
	if to, err = parseOptionalTime(r.URL.Query().Get("to")); err != nil {
		http.Error(w, "invalid to value, expected RFC3339", http.StatusBadRequest)
		return
	}
	beforeCreatedAt, beforeEventID, err := parseBeforeCursor(r.URL.Query().Get("before_cursor"))
	if err != nil {
		http.Error(w, "invalid before_cursor value", http.StatusBadRequest)
		return
	}
	if from != nil && to != nil && from.After(*to) {
		http.Error(w, "from must be before or equal to to", http.StatusBadRequest)
		return
	}

	query := audit.EventQuery{
		CompanyID:       companyID,
		EntityType:      "vault_item",
		EntityID:        itemID,
		Limit:           limit,
		EventType:       strings.TrimSpace(r.URL.Query().Get("event_type")),
		From:            from,
		To:              to,
		BeforeCreatedAt: beforeCreatedAt,
		BeforeEventID:   beforeEventID,
	}
	events := h.Audit.ListEvents(query)
	includeTotal := strings.EqualFold(strings.TrimSpace(r.URL.Query().Get("include_total")), "true")
	var totalCount *int64
	if includeTotal {
		countQuery := query
		countQuery.BeforeCreatedAt = nil
		countQuery.BeforeEventID = ""
		if total, ok := h.Audit.CountEvents(countQuery); ok {
			totalCount = &total
		}
	}
	var nextCursor string
	hasMore := len(events) == limit
	if hasMore {
		nextCursor = encodeBeforeCursor(events[len(events)-1].CreatedAt, events[len(events)-1].ID)
	}

	response := map[string]any{
		"item_id":        itemID,
		"events":         events,
		"returned_count": len(events),
		"has_more":       hasMore,
		"next_cursor":    nextCursor,
	}
	if totalCount != nil {
		response["total_count"] = *totalCount
	}

	writeJSON(w, http.StatusOK, response)
}

func (h VaultHandler) withDefaults() VaultHandler {
	if h.Signer == nil {
		h.Signer = storage.PlaceholderSigner{}
	}
	if h.Reader == nil {
		h.Reader = storage.NoopObjectReader{}
	}
	if h.Extractor == nil {
		h.Extractor = vault.NewSimulatedExtractor()
	}
	if h.Audit == nil {
		h.Audit = audit.NoopLogger{}
	}
	if h.Idempotency == nil {
		h.Idempotency = idempotency.NoopService{}
	}
	return h
}

func extractorName(extractor vault.Extractor) string {
	if extractor == nil {
		return "unknown"
	}
	return reflect.TypeOf(extractor).String()
}

func (h VaultHandler) buildIdempotencyKey(r *http.Request, companyID, userID, operation, resourceID string) (idempotency.Key, bool) {
	key := strings.TrimSpace(r.Header.Get("Idempotency-Key"))
	if key == "" {
		return idempotency.Key{}, false
	}
	return idempotency.Key{
		CompanyID:      companyID,
		UserID:         userID,
		Operation:      operation,
		ResourceID:     resourceID,
		IdempotencyKey: key,
	}, true
}

var errItemNotFound = fmt.Errorf("item not found")
var errItemStatusConflict = fmt.Errorf("item status conflict")

func (h VaultHandler) startProcessing(itemID, companyID, userID, startedEventType string, allowedStatuses []string) (vault.Item, time.Time, error) {
	item, ok := h.Store.StartProcessing(itemID, companyID, allowedStatuses)
	if !ok {
		if _, exists := h.Store.GetByIDForCompany(itemID, companyID); !exists {
			return vault.Item{}, time.Time{}, errItemNotFound
		}
		return vault.Item{}, time.Time{}, errItemStatusConflict
	}

	h.Audit.LogEvent(audit.Event{
		CompanyID:   companyID,
		ActorUserID: userID,
		EventType:   startedEventType,
		EntityType:  "vault_item",
		EntityID:    item.ID,
		Payload: map[string]any{
			"status":    "processing",
			"extractor": extractorName(h.Extractor),
		},
	})

	startedAt := time.Now().UTC()
	go func(currentItem vault.Item) {
		var content []byte
		if h.Reader != nil {
			readContent, readErr := h.Reader.ReadObject(currentItem.StoragePath)
			if readErr != nil && readErr != storage.ErrObjectReaderNotConfigured {
				errorDetails := vault.ClassifyProcessingError("read_object", readErr)
				h.Store.MarkFailed(currentItem.ID, readErr.Error())
				h.Audit.LogEvent(audit.Event{
					CompanyID:   currentItem.CompanyID,
					ActorUserID: currentItem.UploaderUserID,
					EventType:   "vault_item_processing_failed",
					EntityType:  "vault_item",
					EntityID:    currentItem.ID,
					Payload: map[string]any{
						"status":          "failed",
						"error":           readErr.Error(),
						"error_stage":     errorDetails.Stage,
						"error_category":  errorDetails.Category,
						"retryable":       errorDetails.Retryable,
						"extractor":       extractorName(h.Extractor),
						"duration_ms":     time.Since(startedAt).Milliseconds(),
						"content_bytes":   0,
						"document_family": vault.DocumentFamily(currentItem.MimeType),
					},
				})
				h.Metrics.RecordVaultProcessing("failed", vault.DocumentFamily(currentItem.MimeType), errorDetails.Category)
				log.Printf(
					"vault_process_failed item_id=%s stage=%s category=%s retryable=%t error=%s",
					currentItem.ID,
					errorDetails.Stage,
					errorDetails.Category,
					errorDetails.Retryable,
					readErr.Error(),
				)
				return
			}
			content = readContent
		}

		documentType, extractedText, err := h.Extractor.ProcessItem(currentItem, content)
		if err != nil {
			errorDetails := vault.ClassifyProcessingError("extract_content", err)
			h.Store.MarkFailed(currentItem.ID, err.Error())
			h.Audit.LogEvent(audit.Event{
				CompanyID:   currentItem.CompanyID,
				ActorUserID: currentItem.UploaderUserID,
				EventType:   "vault_item_processing_failed",
				EntityType:  "vault_item",
				EntityID:    currentItem.ID,
				Payload: map[string]any{
					"status":          "failed",
					"error":           err.Error(),
					"error_stage":     errorDetails.Stage,
					"error_category":  errorDetails.Category,
					"retryable":       errorDetails.Retryable,
					"extractor":       extractorName(h.Extractor),
					"duration_ms":     time.Since(startedAt).Milliseconds(),
					"content_bytes":   len(content),
					"document_family": vault.DocumentFamily(currentItem.MimeType),
				},
			})
			h.Metrics.RecordVaultProcessing("failed", vault.DocumentFamily(currentItem.MimeType), errorDetails.Category)
			log.Printf(
				"vault_process_failed item_id=%s stage=%s category=%s retryable=%t error=%s",
				currentItem.ID,
				errorDetails.Stage,
				errorDetails.Category,
				errorDetails.Retryable,
				err.Error(),
			)
			return
		}

		h.Store.MarkProcessed(currentItem.ID, documentType, extractedText)
		keyFields := vault.ExtractKeyFields(extractedText)
		missingKeyFields := keyFields.MissingRequired()
		h.Audit.LogEvent(audit.Event{
			CompanyID:   currentItem.CompanyID,
			ActorUserID: currentItem.UploaderUserID,
			EventType:   "vault_item_processed",
			EntityType:  "vault_item",
			EntityID:    currentItem.ID,
			Payload: map[string]any{
				"status":          "processed",
				"document_type":   documentType,
				"extractor":       extractorName(h.Extractor),
				"duration_ms":     time.Since(startedAt).Milliseconds(),
				"content_bytes":   len(content),
				"document_family": vault.DocumentFamily(currentItem.MimeType),
				"key_fields": map[string]any{
					"amount":       keyFields.Amount,
					"date":         keyFields.Date,
					"company_name": keyFields.CompanyName,
				},
				"key_fields_found":   keyFields.FoundCount(),
				"missing_key_fields": missingKeyFields,
			},
		})
		h.Metrics.RecordVaultProcessing("processed", vault.DocumentFamily(currentItem.MimeType), "")
		log.Printf("vault_processed item_id=%s document_type=%s", currentItem.ID, documentType)
	}(item)

	return item, startedAt, nil
}

func parseOptionalTime(rawValue string) (*time.Time, error) {
	value := strings.TrimSpace(rawValue)
	if value == "" {
		return nil, nil
	}
	parsed, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return nil, err
	}
	parsed = parsed.UTC()
	return &parsed, nil
}

func parseBeforeCursor(rawValue string) (*time.Time, string, error) {
	value := strings.TrimSpace(rawValue)
	if value == "" {
		return nil, "", nil
	}

	decoded, err := base64.RawURLEncoding.DecodeString(value)
	if err != nil {
		return nil, "", err
	}
	parts := strings.SplitN(string(decoded), "|", 2)
	if len(parts) != 2 {
		return nil, "", fmt.Errorf("invalid cursor format")
	}

	createdAt, err := time.Parse(time.RFC3339Nano, parts[0])
	if err != nil {
		return nil, "", err
	}
	eventID := strings.TrimSpace(parts[1])
	if eventID == "" {
		return nil, "", fmt.Errorf("missing event id in cursor")
	}

	createdAt = createdAt.UTC()
	return &createdAt, eventID, nil
}

func encodeBeforeCursor(createdAt time.Time, eventID string) string {
	if createdAt.IsZero() || strings.TrimSpace(eventID) == "" {
		return ""
	}
	raw := fmt.Sprintf("%s|%s", createdAt.UTC().Format(time.RFC3339Nano), eventID)
	return base64.RawURLEncoding.EncodeToString([]byte(raw))
}

func newRandomID() string {
	// UUID v4 format to keep compatibility with Postgres uuid columns.
	buf := make([]byte, 16)
	_, _ = rand.Read(buf)
	buf[6] = (buf[6] & 0x0f) | 0x40
	buf[8] = (buf[8] & 0x3f) | 0x80
	hexValue := hex.EncodeToString(buf)
	return fmt.Sprintf("%s-%s-%s-%s-%s",
		hexValue[0:8],
		hexValue[8:12],
		hexValue[12:16],
		hexValue[16:20],
		hexValue[20:32],
	)
}

func sanitizeFileName(fileName string) string {
	ext := filepath.Ext(fileName)
	base := strings.TrimSuffix(fileName, ext)
	base = strings.ToLower(strings.ReplaceAll(base, " ", "_"))
	base = strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '_' || r == '-' {
			return r
		}
		return -1
	}, base)
	if base == "" {
		base = "file"
	}
	return base + strings.ToLower(ext)
}
