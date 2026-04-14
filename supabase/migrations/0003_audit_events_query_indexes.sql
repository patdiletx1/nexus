create index if not exists idx_audit_events_lookup_cursor
  on public.audit_events (company_id, entity_type, entity_id, created_at desc, id desc);

create index if not exists idx_audit_events_lookup_event_type
  on public.audit_events (company_id, entity_type, entity_id, event_type, created_at desc, id desc);
