create table if not exists public.telegram_whitelist (
  id uuid primary key,
  member_id uuid references public.members(id) on delete set null,
  telegram_id bigint,
  telegram_username text not null check (char_length(trim(telegram_username)) > 0),
  display_name text not null check (char_length(trim(display_name)) > 0),
  color_hex text not null check (color_hex ~ '^#[0-9a-f]{6}$'),
  goal text not null default 'other',
  first_name text,
  last_name text,
  last_seen_at timestamptz,
  created_at timestamptz not null default now(),
  updated_at timestamptz not null default now(),
  check (telegram_username = lower(telegram_username)),
  check (telegram_username !~ '^@')
);

create unique index if not exists telegram_whitelist_username_uniq_idx
  on public.telegram_whitelist (telegram_username);

create unique index if not exists telegram_whitelist_telegram_id_uniq_idx
  on public.telegram_whitelist (telegram_id)
  where telegram_id is not null;

create unique index if not exists telegram_whitelist_member_id_uniq_idx
  on public.telegram_whitelist (member_id)
  where member_id is not null;

create index if not exists telegram_whitelist_last_seen_idx
  on public.telegram_whitelist (last_seen_at desc);

alter table public.telegram_whitelist enable row level security;
