create extension if not exists "pgcrypto";

create table if not exists public.members (
  id uuid primary key,
  display_name text not null check (char_length(trim(display_name)) > 0),
  role text not null check (role in ('member', 'admin')),
  created_at timestamptz not null default now()
);

create table if not exists public.connections (
  id uuid primary key,
  member_id uuid not null references public.members(id) on delete cascade,
  goal text not null,
  color_hex text not null check (color_hex ~ '^#[0-9a-f]{6}$'),
  mcp_token text not null unique,
  active boolean not null default true,
  created_at timestamptz not null default now(),
  unique (member_id),
  unique (color_hex)
);

create table if not exists public.vacations (
  id uuid primary key,
  member_id uuid not null references public.members(id) on delete cascade,
  from_date date not null,
  to_date date not null,
  reason text not null default '',
  status text not null check (status in ('pending', 'approved')),
  created_at timestamptz not null default now(),
  updated_at timestamptz not null default now(),
  check (to_date >= from_date)
);

create index if not exists vacations_member_idx on public.vacations(member_id);
create index if not exists vacations_date_range_idx on public.vacations(from_date, to_date);

create table if not exists public.approvals_audit (
  id uuid primary key,
  vacation_id uuid not null references public.vacations(id) on delete cascade,
  approved_by_member_id uuid not null references public.members(id) on delete restrict,
  approved_at timestamptz not null default now()
);

alter table public.members enable row level security;
alter table public.connections enable row level security;
alter table public.vacations enable row level security;
alter table public.approvals_audit enable row level security;

drop policy if exists members_read on public.members;
create policy members_read on public.members for select using (true);

drop policy if exists connections_read on public.connections;
create policy connections_read on public.connections for select using (true);

drop policy if exists vacations_read on public.vacations;
create policy vacations_read on public.vacations for select using (true);

drop policy if exists approvals_audit_read on public.approvals_audit;
create policy approvals_audit_read on public.approvals_audit for select using (true);
