import Link from "next/link";

export default function HomePage() {
  return (
    <main className="grid" style={{ gap: "1rem" }}>
      <section className="card">
        <h1 style={{ marginTop: 0 }}>Планы отпусков команды</h1>
        <p>
          Подключите MCP, выберите свой уникальный цвет и управляйте отпусками из любимого GPT-интерфейса.
        </p>
        <div style={{ display: "flex", gap: "0.75rem" }}>
          <Link href="/connect" className="btn btn-primary" style={{ textDecoration: "none" }}>
            Подключить
          </Link>
          <Link href="/calendar" className="btn" style={{ textDecoration: "none", border: "1px solid #d1d5db" }}>
            Открыть календарь
          </Link>
        </div>
      </section>
    </main>
  );
}
