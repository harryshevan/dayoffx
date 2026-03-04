import "./globals.css";
import Link from "next/link";
import { ReactNode } from "react";

export default function RootLayout({ children }: { children: ReactNode }) {
  return (
    <html lang="ru">
      <body>
        <div className="container">
          <header className="topbar">
            <Link href="/" style={{ textDecoration: "none", fontWeight: 700 }}>
              Team Dayoffs
            </Link>
            <nav style={{ display: "flex", gap: "0.75rem" }}>
              <Link href="/calendar">Календарь</Link>
              <Link href="/connect">Подключить</Link>
            </nav>
          </header>
          {children}
        </div>
      </body>
    </html>
  );
}
