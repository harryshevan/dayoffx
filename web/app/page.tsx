import Link from "next/link";

export default function HomePage() {
  return (
    <main className="grid" style={{ gap: "1rem" }}>
      <section className="card">
        <h1 style={{ marginTop: 0 }}>Team vacation plans</h1>
        <p>
          Connect MCP, pick your unique color, and manage vacations from your favorite GPT interface.
        </p>
        <div style={{ display: "flex", gap: "0.75rem" }}>
          <Link href="/connect" className="btn btn-primary" style={{ textDecoration: "none" }}>
            Connect
          </Link>
          <Link href="/calendar" className="btn" style={{ textDecoration: "none", border: "1px solid #d1d5db" }}>
            Open calendar
          </Link>
        </div>
      </section>
    </main>
  );
}
