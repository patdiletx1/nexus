create extension if not exists pgcrypto;

create table if not exists public.companies (
  id uuid primary key default gen_random_uuid(),
  name text not null,
  business_area text,
  created_at timestamptz not null default now()
);

create table if not exists public.users (
  id uuid primary key references auth.users (id) on delete cascade,
  company_id uuid not null references public.companies (id) on delete restrict,
  role text not null check (role in ('owner', 'admin', 'member')),
  created_at timestamptz not null default now()
);

create table if not exists public.vault_items (
  id uuid primary key default gen_random_uuid(),
  company_id uuid not null references public.companies (id) on delete restrict,
  uploader_user_id uuid references public.users (id) on delete set null,
  storage_path text not null,
  file_name text not null,
  mime_type text not null,
  size_bytes bigint not null check (size_bytes >= 0),
  sha256 text not null,
  status text not null check (status in ('uploaded', 'processing', 'processed', 'failed')),
  created_at timestamptz not null default now(),
  updated_at timestamptz not null default now()
);

create unique index if not exists idx_vault_items_company_sha on public.vault_items (company_id, sha256);

create table if not exists public.audit_events (
  id uuid primary key default gen_random_uuid(),
  company_id uuid not null references public.companies (id) on delete restrict,
  actor_user_id uuid references public.users (id) on delete set null,
  event_type text not null,
  entity_type text not null,
  entity_id uuid,
  payload jsonb not null default '{}'::jsonb,
  created_at timestamptz not null default now()
);

create index if not exists idx_audit_events_company_created_at on public.audit_events (company_id, created_at desc);

create or replace function public.current_company_id()
returns uuid
language sql
stable
security definer
set search_path = public
as $$
  select company_id
  from public.users
  where id = auth.uid();
$$;

create or replace function public.touch_updated_at()
returns trigger
language plpgsql
as $$
begin
  new.updated_at = now();
  return new;
end;
$$;

drop trigger if exists trg_touch_vault_items_updated_at on public.vault_items;
create trigger trg_touch_vault_items_updated_at
before update on public.vault_items
for each row
execute procedure public.touch_updated_at();

alter table public.companies enable row level security;
alter table public.users enable row level security;
alter table public.vault_items enable row level security;
alter table public.audit_events enable row level security;

drop policy if exists companies_select_own on public.companies;
create policy companies_select_own on public.companies
for select
using (id = public.current_company_id());

drop policy if exists users_select_company on public.users;
create policy users_select_company on public.users
for select
using (company_id = public.current_company_id());

drop policy if exists users_update_self on public.users;
create policy users_update_self on public.users
for update
using (id = auth.uid())
with check (id = auth.uid() and company_id = public.current_company_id());

drop policy if exists vault_items_select_company on public.vault_items;
create policy vault_items_select_company on public.vault_items
for select
using (company_id = public.current_company_id());

drop policy if exists vault_items_insert_company on public.vault_items;
create policy vault_items_insert_company on public.vault_items
for insert
with check (
  company_id = public.current_company_id()
  and uploader_user_id = auth.uid()
);

drop policy if exists vault_items_update_company on public.vault_items;
create policy vault_items_update_company on public.vault_items
for update
using (company_id = public.current_company_id())
with check (company_id = public.current_company_id());

drop policy if exists audit_events_select_company on public.audit_events;
create policy audit_events_select_company on public.audit_events
for select
using (company_id = public.current_company_id());

drop policy if exists audit_events_insert_company on public.audit_events;
create policy audit_events_insert_company on public.audit_events
for insert
with check (company_id = public.current_company_id());
