create table if not exists public.calendar_day_overrides (
  date date primary key,
  is_day_off boolean not null,
  reason text not null default '',
  created_by_member_id uuid references public.members(id) on delete set null,
  created_at timestamptz not null default now(),
  updated_at timestamptz not null default now()
);

create index if not exists calendar_day_overrides_dayoff_idx
  on public.calendar_day_overrides (is_day_off);

alter table public.calendar_day_overrides enable row level security;

drop policy if exists calendar_day_overrides_read on public.calendar_day_overrides;
create policy calendar_day_overrides_read on public.calendar_day_overrides for select using (true);
