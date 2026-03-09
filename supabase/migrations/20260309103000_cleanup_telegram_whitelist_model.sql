alter table public.telegram_whitelist
  drop column if exists display_name,
  drop column if exists color_hex,
  drop column if exists goal;
