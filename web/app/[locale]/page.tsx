import { useTranslations } from "next-intl";
import { Link } from "@/i18n/navigation";

export default function HomePage() {
  const t = useTranslations("home");

  return (
    <main className="grid" style={{ gap: "1rem" }}>
      <section className="card">
        <h1 style={{ marginTop: 0 }}>{t("title")}</h1>
        <p>{t("description")}</p>
        <div style={{ display: "flex", gap: "0.75rem" }}>
          <Link href="/calendar" className="btn btn-primary" style={{ textDecoration: "none" }}>
            {t("openCalendar")}
          </Link>
        </div>
        <p style={{ marginBottom: 0, color: "var(--muted)" }}>
          {t("onboardingHint.before")}
          <strong>?</strong>
          {t("onboardingHint.after")}
        </p>
      </section>
    </main>
  );
}
