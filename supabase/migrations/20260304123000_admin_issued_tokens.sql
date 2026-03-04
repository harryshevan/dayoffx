alter table public.connections
  add column if not exists revoked_at timestamptz;

create index if not exists connections_active_token_idx
  on public.connections (mcp_token)
  where active = true;
