create table if not exists public.tender_score_cache (
  id uuid primary key default gen_random_uuid(),
  company_id uuid not null references public.companies (id) on delete cascade,
  tender_external_id text not null,
  profile_fingerprint text not null,
  score integer not null,
  reasons text[] not null default '{}',
  expires_at timestamptz not null,
  created_at timestamptz not null default now(),
  updated_at timestamptz not null default now()
);

create unique index if not exists idx_tender_score_cache_unique
  on public.tender_score_cache (company_id, tender_external_id, profile_fingerprint);

create index if not exists idx_tender_score_cache_expires_at
  on public.tender_score_cache (expires_at);

drop trigger if exists trg_touch_tender_score_cache_updated_at on public.tender_score_cache;
create trigger trg_touch_tender_score_cache_updated_at
before update on public.tender_score_cache
for each row
execute procedure public.touch_updated_at();

alter table public.tender_score_cache enable row level security;

drop policy if exists tender_score_cache_select_company on public.tender_score_cache;
create policy tender_score_cache_select_company on public.tender_score_cache
for select
using (company_id = public.current_company_id());

drop policy if exists tender_score_cache_insert_company on public.tender_score_cache;
create policy tender_score_cache_insert_company on public.tender_score_cache
for insert
with check (company_id = public.current_company_id());

drop policy if exists tender_score_cache_update_company on public.tender_score_cache;
create policy tender_score_cache_update_company on public.tender_score_cache
for update
using (company_id = public.current_company_id())
with check (company_id = public.current_company_id());
