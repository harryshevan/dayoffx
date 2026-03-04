alter table public.connections
  add column if not exists token_id text,
  add column if not exists token_hash text,
  add column if not exists token_version text;

alter table public.connections
  alter column mcp_token drop not null;

alter table public.connections
  drop constraint if exists connections_token_version_check;

alter table public.connections
  add constraint connections_token_version_check
  check (token_version is null or token_version in ('v1'));

create unique index if not exists connections_token_id_uniq_idx
  on public.connections (token_id)
  where token_id is not null;

create index if not exists connections_active_token_id_idx
  on public.connections (token_id)
  where active = true and token_id is not null;
