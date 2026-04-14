create table if not exists public.company_scoring_profiles (
  id uuid primary key default gen_random_uuid(),
  company_id uuid not null unique references public.companies (id) on delete cascade,
  preferred_region text,
  keywords text[] not null default '{}',
  updated_by_user_id uuid references public.users (id) on delete set null,
  created_at timestamptz not null default now(),
  updated_at timestamptz not null default now()
);

drop trigger if exists trg_touch_company_scoring_profiles_updated_at on public.company_scoring_profiles;
create trigger trg_touch_company_scoring_profiles_updated_at
before update on public.company_scoring_profiles
for each row
execute procedure public.touch_updated_at();

alter table public.company_scoring_profiles enable row level security;

drop policy if exists company_scoring_profiles_select_company on public.company_scoring_profiles;
create policy company_scoring_profiles_select_company on public.company_scoring_profiles
for select
using (company_id = public.current_company_id());

drop policy if exists company_scoring_profiles_insert_company on public.company_scoring_profiles;
create policy company_scoring_profiles_insert_company on public.company_scoring_profiles
for insert
with check (company_id = public.current_company_id());

drop policy if exists company_scoring_profiles_update_company on public.company_scoring_profiles;
create policy company_scoring_profiles_update_company on public.company_scoring_profiles
for update
using (company_id = public.current_company_id())
with check (company_id = public.current_company_id());
