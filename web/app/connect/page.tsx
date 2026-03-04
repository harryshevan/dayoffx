"use client";

import { FormEvent, useEffect, useMemo, useRef, useState } from "react";
import { connectMember } from "@/lib/api";
import { Goal, MemberConnection } from "@/lib/types";

const MCP_SERVER_NAME = "dayoff-mcp";

const GOAL_INSTRUCTIONS: Record<Goal, string> = {
  cursor: "Copy the JSON and paste it into Cursor MCP settings (mcp.json).",
  claude_desktop: "Copy the MCP URL and paste it into your Claude Desktop config (mcpServers).",
  other: "Use the MCP URL and token in any client that supports Model Context Protocol."
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
  const [copyError, setCopyError] = useState<string | null>(null);
  const [copiedKey, setCopiedKey] = useState<string | null>(null);
  const [submitting, setSubmitting] = useState(false);
  const copiedTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null);

  const instruction = useMemo(() => GOAL_INSTRUCTIONS[goal], [goal]);

  async function copyText(value: string) {
    await navigator.clipboard.writeText(value);
  }

  useEffect(() => {
    return () => {
      if (copiedTimerRef.current) {
        clearTimeout(copiedTimerRef.current);
      }
    };
  }, []);

  async function copyWithStatus(value: string, key: string) {
    try {
      await copyText(value);
      setCopyError(null);
      setCopiedKey(key);
      if (copiedTimerRef.current) {
        clearTimeout(copiedTimerRef.current);
      }
      copiedTimerRef.current = setTimeout(() => {
        setCopiedKey((prevKey) => (prevKey === key ? null : prevKey));
      }, 2000);
    } catch {
      setCopyError("Failed to copy. Please copy manually.");
    }
  }

  const mcpServerUrl = connection?.mcpServerUrl || "";
  const serverBlockJson = useMemo(() => {
    if (!connection) {
      return "";
    }
    return JSON.stringify(
      {
        [MCP_SERVER_NAME]: {
          url: mcpServerUrl || "https://YOUR_MCP_SERVER_URL/mcp",
          headers: {
            Authorization: `Bearer ${connection.mcpToken}`
          }
        }
      },
      null,
      2
    );
  }, [connection, mcpServerUrl]);

  const fullMcpJson = useMemo(() => {
    if (!connection) {
      return "";
    }
    return JSON.stringify(
      {
        mcpServers: JSON.parse(serverBlockJson)
      },
      null,
      2
    );
  }, [connection, serverBlockJson]);

  async function onSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setSubmitting(true);
    setError(null);
    setCopyError(null);
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
        <h1 style={{ marginTop: 0 }}>Connect to MCP</h1>
        <p style={{ marginTop: 0 }}>
          One member = one MCP connection = one unique color on the calendar.
        </p>

        <form className="grid" onSubmit={onSubmit}>
          <label className="grid" style={{ gap: "0.4rem" }}>
            Team member name
            <input value={displayName} onChange={(event) => setDisplayName(event.target.value)} required />
          </label>

          <label className="grid" style={{ gap: "0.4rem" }}>
            Connection goal
            <select value={goal} onChange={(event) => setGoal(event.target.value as Goal)}>
              <option value="cursor">Cursor</option>
              <option value="claude_desktop">Claude Desktop</option>
              <option value="other">Other</option>
            </select>
          </label>

          <div className="card" style={{ padding: "0.7rem" }}>
            <strong>1-click instruction:</strong>
            <div style={{ marginTop: "0.25rem", fontSize: "0.9rem" }}>{instruction}</div>
          </div>

          <div className="grid" style={{ gap: "0.4rem" }}>
            <span>Pick a unique color</span>
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
            Admin secret (optional)
            <input
              value={adminSecret}
              onChange={(event) => setAdminSecret(event.target.value)}
              placeholder="Admins only"
            />
          </label>

          <button className="btn btn-primary" disabled={submitting}>
            {submitting ? "Connecting..." : "Connect"}
          </button>
        </form>

        {error ? <p style={{ color: "#dc2626" }}>{error}</p> : null}
        {copyError ? <p style={{ color: "#dc2626", margin: 0 }}>{copyError}</p> : null}

        {connection ? (
          <div className="card" style={{ marginTop: "0.75rem" }}>
            <strong>Connection ready</strong>
            <div>ID: {connection.memberId}</div>
            <div>Role: {connection.role}</div>
            <div style={{ display: "flex", gap: "0.4rem", alignItems: "center", flexWrap: "wrap" }}>
              MCP URL: {connection.mcpServerUrl || "set MCP_SERVER_URL in API env"}
              {connection.mcpServerUrl ? (
                <CopyIconButton
                  copied={copiedKey === "mcp-url"}
                  onClick={() => copyWithStatus(connection.mcpServerUrl || "", "mcp-url")}
                  label="Copy MCP URL"
                />
              ) : null}
            </div>
            <div style={{ display: "flex", gap: "0.4rem", alignItems: "center", flexWrap: "wrap" }}>
              MCP token: {connection.mcpToken}
              <CopyIconButton
                copied={copiedKey === "mcp-token"}
                onClick={() => copyWithStatus(connection.mcpToken, "mcp-token")}
                label="Copy MCP token"
              />
            </div>
            <div className="grid" style={{ marginTop: "0.75rem" }}>
              <strong>Ready JSON for Cursor</strong>
              {!mcpServerUrl ? (
                <div style={{ color: "#b45309" }}>
                  `MCP_SERVER_URL` is not configured in the API. Fill in the URL manually before saving.
                </div>
              ) : null}
              <div style={{ display: "flex", gap: "0.5rem", flexWrap: "wrap" }}>
                <div style={{ display: "inline-flex", gap: "0.4rem", alignItems: "center" }}>
                  Full mcp.json
                  <CopyIconButton
                    copied={copiedKey === "full-mcp-json"}
                    onClick={() => copyWithStatus(fullMcpJson, "full-mcp-json")}
                    label="Copy full mcp.json"
                  />
                </div>
                <div style={{ display: "inline-flex", gap: "0.4rem", alignItems: "center" }}>
                  Server block only
                  <CopyIconButton
                    copied={copiedKey === "server-block"}
                    onClick={() => copyWithStatus(serverBlockJson, "server-block")}
                    label="Copy server block only"
                  />
                </div>
              </div>
              <label className="grid" style={{ gap: "0.35rem" }}>
                Full mcp.json
                <textarea
                  readOnly
                  value={fullMcpJson}
                  style={{ width: "100%", minHeight: 160, fontFamily: "ui-monospace, SFMono-Regular, Menlo, monospace" }}
                />
              </label>
              <label className="grid" style={{ gap: "0.35rem" }}>
                Server block only
                <textarea
                  readOnly
                  value={serverBlockJson}
                  style={{ width: "100%", minHeight: 120, fontFamily: "ui-monospace, SFMono-Regular, Menlo, monospace" }}
                />
              </label>
            </div>
          </div>
        ) : null}
      </section>
    </main>
  );
}

function CopyIconButton({
  copied,
  onClick,
  label
}: {
  copied: boolean;
  onClick: () => void;
  label: string;
}) {
  return (
    <button
      type="button"
      className="btn"
      onClick={onClick}
      title={copied ? "Copied" : label}
      aria-label={copied ? "Copied" : label}
      style={{
        display: "inline-flex",
        alignItems: "center",
        justifyContent: "center",
        width: 34,
        height: 34,
        padding: 0
      }}
    >
      {copied ? <CheckIcon /> : <CopyIcon />}
    </button>
  );
}

function CopyIcon() {
  return (
    <svg viewBox="0 0 24 24" width="18" height="18" fill="none" stroke="currentColor" strokeWidth="2">
      <rect x="9" y="9" width="10" height="10" rx="2" />
      <path d="M5 15V7a2 2 0 0 1 2-2h8" />
    </svg>
  );
}

function CheckIcon() {
  return (
    <svg viewBox="0 0 24 24" width="18" height="18" fill="none" stroke="currentColor" strokeWidth="2">
      <path d="m5 13 4 4L19 7" />
    </svg>
  );
}
