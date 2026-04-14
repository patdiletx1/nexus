package vault

import (
	"context"
	"log"
	"time"

	"github.com/jackc/pgx/v4/pgxpool"
)

type PostgresStore struct {
	pool *pgxpool.Pool
}

func NewPostgresStore(pool *pgxpool.Pool) *PostgresStore {
	return &PostgresStore{pool: pool}
}

func (s *PostgresStore) Save(item Item) {
	if item.UpdatedAt.IsZero() {
		item.UpdatedAt = item.CreatedAt
	}
	item.CreatedAt = ensureUTC(item.CreatedAt)
	item.UpdatedAt = ensureUTC(item.UpdatedAt)

	_, err := s.pool.Exec(context.Background(), `
		insert into public.vault_items (
			id, company_id, uploader_user_id, storage_path, file_name, mime_type, size_bytes,
			sha256, status, document_type, extracted_text, error_message, created_at, updated_at
		) values (
			$1, $2, $3, $4, $5, $6, $7,
			$8, $9, $10, $11, $12, $13, $14
		)
	`,
		item.ID, item.CompanyID, item.UploaderUserID, item.StoragePath, item.FileName, item.MimeType, item.SizeBytes,
		item.SHA256, item.Status, nullIfEmpty(item.DocumentType), nullIfEmpty(item.ExtractedText), nullIfEmpty(item.ErrorMessage), item.CreatedAt, item.UpdatedAt,
	)
	if err != nil {
		log.Printf("vault_store_save_failed item_id=%s error=%v", item.ID, err)
	}
}

func (s *PostgresStore) GetByIDForCompany(itemID, companyID string) (Item, bool) {
	row := s.pool.QueryRow(context.Background(), `
		select id, company_id, uploader_user_id, storage_path, file_name, mime_type, size_bytes,
		       sha256, status, document_type, extracted_text, error_message, created_at, updated_at
		from public.vault_items
		where id = $1 and company_id = $2
	`, itemID, companyID)

	item, err := scanItem(row.Scan)
	if err != nil {
		return Item{}, false
	}
	return item, true
}

func (s *PostgresStore) StartProcessing(itemID, companyID string, allowedStatuses []string) (Item, bool) {
	if len(allowedStatuses) == 0 {
		return Item{}, false
	}

	row := s.pool.QueryRow(context.Background(), `
		update public.vault_items
		set status = 'processing', error_message = null, updated_at = now()
		where id = $1 and company_id = $2 and status = any($3::text[])
		returning id, company_id, uploader_user_id, storage_path, file_name, mime_type, size_bytes,
		          sha256, status, document_type, extracted_text, error_message, created_at, updated_at
	`, itemID, companyID, allowedStatuses)

	item, err := scanItem(row.Scan)
	if err != nil {
		return Item{}, false
	}
	return item, true
}

func (s *PostgresStore) MarkProcessed(itemID, documentType, extractedText string) (Item, bool) {
	row := s.pool.QueryRow(context.Background(), `
		update public.vault_items
		set status = 'processed',
		    document_type = $2,
		    extracted_text = $3,
		    error_message = null,
		    updated_at = now()
		where id = $1
		returning id, company_id, uploader_user_id, storage_path, file_name, mime_type, size_bytes,
		          sha256, status, document_type, extracted_text, error_message, created_at, updated_at
	`, itemID, nullIfEmpty(documentType), nullIfEmpty(extractedText))

	item, err := scanItem(row.Scan)
	if err != nil {
		return Item{}, false
	}
	return item, true
}

func (s *PostgresStore) MarkFailed(itemID, errMessage string) (Item, bool) {
	row := s.pool.QueryRow(context.Background(), `
		update public.vault_items
		set status = 'failed', error_message = $2, updated_at = now()
		where id = $1
		returning id, company_id, uploader_user_id, storage_path, file_name, mime_type, size_bytes,
		          sha256, status, document_type, extracted_text, error_message, created_at, updated_at
	`, itemID, nullIfEmpty(errMessage))

	item, err := scanItem(row.Scan)
	if err != nil {
		return Item{}, false
	}
	return item, true
}

func (s *PostgresStore) ListByCompany(companyID string) []Item {
	rows, err := s.pool.Query(context.Background(), `
		select id, company_id, uploader_user_id, storage_path, file_name, mime_type, size_bytes,
		       sha256, status, document_type, extracted_text, error_message, created_at, updated_at
		from public.vault_items
		where company_id = $1
		order by created_at desc
	`, companyID)
	if err != nil {
		log.Printf("vault_store_list_failed company_id=%s error=%v", companyID, err)
		return []Item{}
	}
	defer rows.Close()

	items := make([]Item, 0)
	for rows.Next() {
		item, err := scanItem(rows.Scan)
		if err != nil {
			log.Printf("vault_store_list_scan_failed company_id=%s error=%v", companyID, err)
			continue
		}
		items = append(items, item)
	}
	return items
}

func scanItem(scan func(dest ...any) error) (Item, error) {
	var item Item
	var documentType *string
	var extractedText *string
	var errorMessage *string
	var uploaderUserID *string

	err := scan(
		&item.ID,
		&item.CompanyID,
		&uploaderUserID,
		&item.StoragePath,
		&item.FileName,
		&item.MimeType,
		&item.SizeBytes,
		&item.SHA256,
		&item.Status,
		&documentType,
		&extractedText,
		&errorMessage,
		&item.CreatedAt,
		&item.UpdatedAt,
	)
	if err != nil {
		return Item{}, err
	}
	if uploaderUserID != nil {
		item.UploaderUserID = *uploaderUserID
	}
	if documentType != nil {
		item.DocumentType = *documentType
	}
	if extractedText != nil {
		item.ExtractedText = *extractedText
	}
	if errorMessage != nil {
		item.ErrorMessage = *errorMessage
	}

	return item, nil
}

func nullIfEmpty(value string) any {
	if value == "" {
		return nil
	}
	return value
}

func ensureUTC(value time.Time) time.Time {
	if value.IsZero() {
		return time.Now().UTC()
	}
	return value.UTC()
}
