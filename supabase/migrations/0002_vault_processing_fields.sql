alter table if exists public.vault_items
  add column if not exists document_type text,
  add column if not exists extracted_text text,
  add column if not exists error_message text;
