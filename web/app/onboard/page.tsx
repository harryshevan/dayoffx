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
  return (
    <main className="grid" style={{ gap: "1rem" }}>
      <section className="card">
        <h1 style={{ marginTop: 0, marginBottom: "0.5rem" }}>How to connect</h1>
        <p style={{ marginTop: 0, marginBottom: "0.75rem" }}>
          This page shows the connection format only. Real credentials are issued by an admin through MCP tool
          <code> createUser</code>.
        </p>
        <div className="card" style={{ background: "#f8fafc" }}>
          <strong>Important</strong>
          <div style={{ marginTop: "0.35rem" }}>
            The token below is fictional and will never work in production: <code>{EXAMPLE_TOKEN}</code>
          </div>
        </div>
      </section>

      <section className="card">
        <h2 style={{ marginTop: 0, marginBottom: "0.6rem" }}>How to connect</h2>
        <ol style={{ margin: 0, paddingInlineStart: "1.1rem", display: "grid", gap: "0.45rem" }}>
          <li>Ask an admin to create your user with MCP tool <code>createUser</code>.</li>
          <li>Receive your personal MCP token from the admin.</li>
          <li>Open Cursor MCP settings and paste config similar to the example below.</li>
          <li>Replace both URL and token with real values from your admin.</li>
        </ol>
      </section>

      <section className="card">
        <h2 style={{ marginTop: 0, marginBottom: "0.6rem" }}>Cursor mcp.json example</h2>
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
