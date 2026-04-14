create table if not exists public.idempotency_keys (
  id uuid primary key default gen_random_uuid(),
  company_id uuid not null references public.companies (id) on delete cascade,
  user_id uuid references public.users (id) on delete set null,
  operation text not null,
  resource_id text not null,
  idempotency_key text not null,
  status_code integer not null,
  response_payload jsonb not null default '{}'::jsonb,
  created_at timestamptz not null default now(),
  expires_at timestamptz not null
);

create unique index if not exists idx_idempotency_keys_unique
  on public.idempotency_keys (company_id, user_id, operation, resource_id, idempotency_key);

create index if not exists idx_idempotency_keys_expires_at
  on public.idempotency_keys (expires_at);

alter table public.idempotency_keys enable row level security;

drop policy if exists idempotency_keys_select_company on public.idempotency_keys;
create policy idempotency_keys_select_company on public.idempotency_keys
for select
using (company_id = public.current_company_id());

drop policy if exists idempotency_keys_insert_company on public.idempotency_keys;
create policy idempotency_keys_insert_company on public.idempotency_keys
for insert
with check (company_id = public.current_company_id());
