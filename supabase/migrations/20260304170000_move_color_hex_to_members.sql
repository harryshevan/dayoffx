alter table public.members
  add column if not exists color_hex text;

update public.members as m
set color_hex = c.color_hex
from public.connections as c
where c.member_id = m.id
  and m.color_hex is null;

do $$
begin
  if exists (
    select 1
    from public.members
    where color_hex is null
  ) then
    raise exception 'cannot enforce members.color_hex not null: null values remain after backfill';
  end if;
end $$;

alter table public.members
  alter column color_hex set not null;

alter table public.members
  drop constraint if exists members_color_hex_check;

alter table public.members
  add constraint members_color_hex_check
  check (color_hex ~ '^#[0-9a-f]{6}$');

alter table public.connections
  drop column if exists color_hex;
