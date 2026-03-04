"use client";

import { FormEvent, useMemo, useState } from "react";
import { connectMember } from "@/lib/api";
import { Goal, MemberConnection } from "@/lib/types";

const GOAL_INSTRUCTIONS: Record<Goal, string> = {
  cursor: "Скопируйте MCP URL и добавьте его в Cursor -> Settings -> MCP -> Add server.",
  claude_desktop: "Скопируйте MCP URL и вставьте его в конфиг Claude Desktop (mcpServers).",
  other: "Используйте MCP URL и token в любом клиенте с поддержкой Model Context Protocol."
};

const PALETTE = [
  "#ef4444",
  "#f97316",
  "#eab308",
  "#22c55e",
  "#14b8a6",
  "#3b82f6",
  "#6366f1",
  "#a855f7",
  "#ec4899"
];

export default function ConnectPage() {
  const [displayName, setDisplayName] = useState("");
  const [goal, setGoal] = useState<Goal>("cursor");
  const [colorHex, setColorHex] = useState(PALETTE[0]);
  const [adminSecret, setAdminSecret] = useState("");
  const [connection, setConnection] = useState<MemberConnection | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [submitting, setSubmitting] = useState(false);

  const instruction = useMemo(() => GOAL_INSTRUCTIONS[goal], [goal]);

  async function copyText(value: string) {
    await navigator.clipboard.writeText(value);
  }

  async function onSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setSubmitting(true);
    setError(null);
    setConnection(null);

    try {
      const result = await connectMember({
        displayName,
        goal,
        colorHex,
        adminSecret: adminSecret.trim() || undefined
      });
      setConnection(result);
    } catch (submitError) {
      setError(submitError instanceof Error ? submitError.message : "Connect failed");
    } finally {
      setSubmitting(false);
    }
  }

  return (
    <main className="grid">
      <section className="card">
        <h1 style={{ marginTop: 0 }}>Подключение к MCP</h1>
        <p style={{ marginTop: 0 }}>
          Один участник = один MCP connection = один уникальный цвет в календаре.
        </p>

        <form className="grid" onSubmit={onSubmit}>
          <label className="grid" style={{ gap: "0.4rem" }}>
            Имя в команде
            <input value={displayName} onChange={(event) => setDisplayName(event.target.value)} required />
          </label>

          <label className="grid" style={{ gap: "0.4rem" }}>
            Цель подключения
            <select value={goal} onChange={(event) => setGoal(event.target.value as Goal)}>
              <option value="cursor">Cursor</option>
              <option value="claude_desktop">Claude Desktop</option>
              <option value="other">Other</option>
            </select>
          </label>

          <div className="card" style={{ padding: "0.7rem" }}>
            <strong>Инструкция в 1 клик:</strong>
            <div style={{ marginTop: "0.25rem", fontSize: "0.9rem" }}>{instruction}</div>
          </div>

          <div className="grid" style={{ gap: "0.4rem" }}>
            <span>Выберите уникальный цвет</span>
            <div style={{ display: "flex", gap: "0.5rem", flexWrap: "wrap" }}>
              {PALETTE.map((paletteColor) => (
                <button
                  key={paletteColor}
                  type="button"
                  onClick={() => setColorHex(paletteColor)}
                  className="btn"
                  style={{
                    width: 34,
                    height: 34,
                    padding: 0,
                    border: colorHex === paletteColor ? "2px solid #111827" : "1px solid #d1d5db",
                    background: paletteColor
                  }}
                  title={paletteColor}
                />
              ))}
            </div>
          </div>

          <label className="grid" style={{ gap: "0.4rem" }}>
            Admin secret (опционально)
            <input
              value={adminSecret}
              onChange={(event) => setAdminSecret(event.target.value)}
              placeholder="Только для администраторов"
            />
          </label>

          <button className="btn btn-primary" disabled={submitting}>
            {submitting ? "Подключаем..." : "Подключить"}
          </button>
        </form>

        {error ? <p style={{ color: "#dc2626" }}>{error}</p> : null}

        {connection ? (
          <div className="card" style={{ marginTop: "0.75rem" }}>
            <strong>Подключение готово</strong>
            <div>ID: {connection.memberId}</div>
            <div>Роль: {connection.role}</div>
            <div style={{ display: "flex", gap: "0.4rem", alignItems: "center", flexWrap: "wrap" }}>
              MCP URL: {connection.mcpServerUrl || "set MCP_SERVER_URL in API env"}
              {connection.mcpServerUrl ? (
                <button type="button" className="btn" onClick={() => copyText(connection.mcpServerUrl || "")}>
                  Скопировать URL
                </button>
              ) : null}
            </div>
            <div style={{ display: "flex", gap: "0.4rem", alignItems: "center", flexWrap: "wrap" }}>
              MCP token: {connection.mcpToken}
              <button type="button" className="btn" onClick={() => copyText(connection.mcpToken)}>
                Скопировать token
              </button>
            </div>
          </div>
        ) : null}
      </section>
    </main>
  );
}
