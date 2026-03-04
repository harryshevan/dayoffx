drop index if exists public.connections_active_token_idx;

alter table public.connections
  drop column if exists mcp_token;
