"use client";

import { useLocale, useTranslations } from "next-intl";
import { Link, usePathname } from "@/i18n/navigation";

export function LocaleSwitcher() {
  const locale = useLocale();
  const pathname = usePathname();
  const t = useTranslations("common");
  const nextLocale = locale === "ru" ? "en" : "ru";

  return (
    <Link
      href={pathname}
      locale={nextLocale}
      className="locale-flag"
      aria-label={t("switchLanguageTo", { locale: nextLocale.toUpperCase() })}
      title={t("switchLanguageTo", { locale: nextLocale.toUpperCase() })}
    >
      {locale}
    </Link>
  );
}
