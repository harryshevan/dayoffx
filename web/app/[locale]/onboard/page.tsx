import { useTranslations } from "next-intl";

const EXAMPLE_TOKEN = "dayoff_example_token_do_not_use";
const EXAMPLE_MCP_URL = "https://your-dayoff-host.example.com/mcp";

const cursorConfigExample = `{
  "mcpServers": {
    "dayoff-mcp": {
      "url": "${EXAMPLE_MCP_URL}",
      "headers": {
        "Authorization": "Bearer ${EXAMPLE_TOKEN}"
      }
    }
  }
}`;

export default function OnboardPage() {
  const t = useTranslations("onboard");

  return (
    <main className="grid" style={{ gap: "1rem" }}>
      <section className="card">
        <h1 style={{ marginTop: 0, marginBottom: "0.5rem" }}>{t("title")}</h1>
        <p style={{ marginTop: 0, marginBottom: "0.75rem" }}>
          {t("intro.beforeCode")}
          <code> admin.user.create</code>
          {t("intro.afterCode")}
        </p>
        <div className="card" style={{ background: "var(--surface-soft)", borderColor: "var(--border-strong)" }}>
          <strong>{t("important.title")}</strong>
          <div style={{ marginTop: "0.35rem" }}>
            {t("important.text")} <code>{EXAMPLE_TOKEN}</code>
          </div>
        </div>
      </section>

      <section className="card">
        <h2 style={{ marginTop: 0, marginBottom: "0.6rem" }}>{t("stepsTitle")}</h2>
        <ol style={{ margin: 0, paddingInlineStart: "1.1rem", display: "grid", gap: "0.45rem" }}>
          <li>
            {t("steps.step1.beforeCode")}
            <code>admin.user.create</code>
            {t("steps.step1.afterCode")}
          </li>
          <li>{t("steps.step2")}</li>
          <li>{t("steps.step3")}</li>
          <li>{t("steps.step4")}</li>
        </ol>
      </section>

      <section className="card">
        <h2 style={{ marginTop: 0, marginBottom: "0.6rem" }}>{t("configTitle")}</h2>
        <textarea
          readOnly
          value={cursorConfigExample}
          style={{
            width: "100%",
            minHeight: 220,
            fontFamily: "ui-monospace, SFMono-Regular, Menlo, monospace",
            background: "#0f172a",
            color: "#e2e8f0",
            border: "1px solid #1e293b",
            borderRadius: 10,
            padding: "0.75rem"
          }}
        />
      </section>
    </main>
  );
}
