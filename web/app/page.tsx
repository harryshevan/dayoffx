import Link from "next/link";

export default function HomePage() {
  return (
    <main className="grid" style={{ gap: "1rem" }}>
      <section className="card">
        <h1 style={{ marginTop: 0 }}>Team vacation plans</h1>
        <p>
          Connect to team MCP, then manage vacations from your favorite GPT interface.
        </p>
        <div style={{ display: "flex", gap: "0.75rem" }}>
          <Link href="/calendar" className="btn btn-primary" style={{ textDecoration: "none" }}>
            Open calendar
          </Link>
        </div>
        <p style={{ marginBottom: 0, color: "var(--muted)" }}>
          Need a quick guide? Open onboarding with the <strong>?</strong> button in the bottom-left corner.
        </p>
      </section>
    </main>
  );
}
