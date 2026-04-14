package vault

import (
	"sort"
	"sync"
	"time"
)

type Item struct {
	ID             string    `json:"id"`
	CompanyID      string    `json:"company_id"`
	UploaderUserID string    `json:"uploader_user_id"`
	StoragePath    string    `json:"storage_path"`
	FileName       string    `json:"file_name"`
	MimeType       string    `json:"mime_type"`
	SizeBytes      int64     `json:"size_bytes"`
	SHA256         string    `json:"sha256"`
	Status         string    `json:"status"`
	DocumentType   string    `json:"document_type,omitempty"`
	ExtractedText  string    `json:"extracted_text,omitempty"`
	ErrorMessage   string    `json:"error_message,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type Store interface {
	Save(item Item)
	GetByIDForCompany(itemID, companyID string) (Item, bool)
	StartProcessing(itemID, companyID string, allowedStatuses []string) (Item, bool)
	MarkProcessed(itemID, documentType, extractedText string) (Item, bool)
	MarkFailed(itemID, errMessage string) (Item, bool)
	ListByCompany(companyID string) []Item
}

type InMemoryStore struct {
	mu    sync.RWMutex
	items map[string]Item
}

func NewInMemoryStore() *InMemoryStore {
	return &InMemoryStore{
		items: map[string]Item{},
	}
}

func (s *InMemoryStore) Save(item Item) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if item.UpdatedAt.IsZero() {
		item.UpdatedAt = item.CreatedAt
	}
	s.items[item.ID] = item
}

func (s *InMemoryStore) GetByIDForCompany(itemID, companyID string) (Item, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	item, ok := s.items[itemID]
	if !ok || item.CompanyID != companyID {
		return Item{}, false
	}

	return item, true
}

func (s *InMemoryStore) StartProcessing(itemID, companyID string, allowedStatuses []string) (Item, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	item, ok := s.items[itemID]
	if !ok || item.CompanyID != companyID {
		return Item{}, false
	}
	if !containsStatus(allowedStatuses, item.Status) {
		return Item{}, false
	}

	item.Status = "processing"
	item.ErrorMessage = ""
	item.UpdatedAt = time.Now().UTC()
	s.items[item.ID] = item

	return item, true
}

func containsStatus(statuses []string, status string) bool {
	for _, candidate := range statuses {
		if candidate == status {
			return true
		}
	}
	return false
}

func (s *InMemoryStore) MarkProcessed(itemID, documentType, extractedText string) (Item, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	item, ok := s.items[itemID]
	if !ok {
		return Item{}, false
	}

	item.Status = "processed"
	item.DocumentType = documentType
	item.ExtractedText = extractedText
	item.ErrorMessage = ""
	item.UpdatedAt = time.Now().UTC()
	s.items[item.ID] = item

	return item, true
}

func (s *InMemoryStore) MarkFailed(itemID, errMessage string) (Item, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	item, ok := s.items[itemID]
	if !ok {
		return Item{}, false
	}

	item.Status = "failed"
	item.ErrorMessage = errMessage
	item.UpdatedAt = time.Now().UTC()
	s.items[item.ID] = item

	return item, true
}

func (s *InMemoryStore) ListByCompany(companyID string) []Item {
	s.mu.RLock()
	defer s.mu.RUnlock()

	out := make([]Item, 0)
	for _, item := range s.items {
		if item.CompanyID == companyID {
			out = append(out, item)
		}
	}

	sort.Slice(out, func(i, j int) bool {
		return out[i].CreatedAt.After(out[j].CreatedAt)
	})

	return out
}
