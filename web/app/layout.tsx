import "./globals.css";
import Link from "next/link";
import { ReactNode } from "react";
import { ThemeToggle } from "@/components/theme-toggle";
import { OnboardingWidget } from "@/components/onboarding-widget";

const themeInitScript = `
(() => {
  const storageKey = "dayoffs-theme";
  const savedTheme = window.localStorage.getItem(storageKey);
  const preferredTheme = window.matchMedia("(prefers-color-scheme: dark)").matches ? "dark" : "light";
  const theme = savedTheme === "dark" || savedTheme === "light" ? savedTheme : preferredTheme;
  document.documentElement.dataset.theme = theme;
  document.documentElement.style.colorScheme = theme;
})();
`;

export default function RootLayout({ children }: { children: ReactNode }) {
  return (
    <html lang="en" suppressHydrationWarning>
      <head>
        <script dangerouslySetInnerHTML={{ __html: themeInitScript }} />
      </head>
      <body>
        <div className="container">
          <header className="topbar">
            <Link href="/" style={{ textDecoration: "none", fontWeight: 700 }}>
              Team Dayoffs
            </Link>
            <div className="topbar-actions">
              <nav className="topnav">
                <Link href="/calendar">Calendar</Link>
              </nav>
              <ThemeToggle />
            </div>
          </header>
          {children}
        </div>
        <OnboardingWidget />
      </body>
    </html>
  );
}
