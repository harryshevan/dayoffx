import { ReactNode } from "react";
import { hasLocale, NextIntlClientProvider } from "next-intl";
import { getMessages, setRequestLocale } from "next-intl/server";
import { notFound } from "next/navigation";
import { ThemeToggle } from "@/components/theme-toggle";
import { OnboardingWidget } from "@/components/onboarding-widget";
import { LocaleSwitcher } from "@/components/locale-switcher";
import { routing } from "@/i18n/routing";

type LocaleLayoutProps = {
  children: ReactNode;
  params: Promise<{ locale: string }>;
};

function TopBar() {
  return (
    <header className="topbar">
      <LocaleSwitcher />
      <div className="topbar-actions">
        <ThemeToggle />
      </div>
    </header>
  );
}

export default async function LocaleLayout({ children, params }: LocaleLayoutProps) {
  const { locale } = await params;
  if (!hasLocale(routing.locales, locale)) {
    notFound();
  }

  setRequestLocale(locale);
  const messages = await getMessages();

  return (
    <NextIntlClientProvider locale={locale} messages={messages}>
      <div className="container">
        <TopBar />
        {children}
      </div>
      <footer className="site-credit">
        dayoffs by{" "}
        <a href="https://t.me/xigax" target="_blank" rel="noopener noreferrer">
          @xigax
        </a>
      </footer>
      <OnboardingWidget />
    </NextIntlClientProvider>
  );
}
