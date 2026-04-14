create table if not exists public.tenders (
  id uuid primary key default gen_random_uuid(),
  company_id uuid not null references public.companies (id) on delete cascade,
  external_id text not null,
  title text not null,
  description text,
  region text,
  closing_at timestamptz,
  published_at timestamptz,
  source text not null default 'chilecompra',
  source_payload jsonb not null default '{}'::jsonb,
  last_synced_at timestamptz not null default now(),
  created_at timestamptz not null default now(),
  updated_at timestamptz not null default now()
);

create unique index if not exists idx_tenders_company_external
  on public.tenders (company_id, external_id);

create index if not exists idx_tenders_company_closing
  on public.tenders (company_id, closing_at desc);

drop trigger if exists trg_touch_tenders_updated_at on public.tenders;
create trigger trg_touch_tenders_updated_at
before update on public.tenders
for each row
execute procedure public.touch_updated_at();

alter table public.tenders enable row level security;

drop policy if exists tenders_select_company on public.tenders;
create policy tenders_select_company on public.tenders
for select
using (company_id = public.current_company_id());

drop policy if exists tenders_insert_company on public.tenders;
create policy tenders_insert_company on public.tenders
for insert
with check (company_id = public.current_company_id());

drop policy if exists tenders_update_company on public.tenders;
create policy tenders_update_company on public.tenders
for update
using (company_id = public.current_company_id())
with check (company_id = public.current_company_id());
