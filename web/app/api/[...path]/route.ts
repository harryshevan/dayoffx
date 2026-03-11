import { NextRequest, NextResponse } from "next/server";

function getUpstreamBaseUrl(): string {
  const configured = process.env.API_INTERNAL_URL?.trim();
  if (configured) {
    return configured;
  }
  if (process.env.NODE_ENV === "development") {
    return "http://localhost:8080";
  }
  throw new Error("API_INTERNAL_URL is not set");
}

async function proxyRequest(request: NextRequest): Promise<Response> {
  const upstreamBaseUrl = getUpstreamBaseUrl();
  const upstreamPath = request.nextUrl.pathname.replace(/^\/api\/?/, "");
  const upstreamUrl = new URL(`/${upstreamPath}`, upstreamBaseUrl);
  upstreamUrl.search = request.nextUrl.search;

  const headers = new Headers(request.headers);
  headers.delete("host");
  headers.delete("connection");
  headers.delete("content-length");

  let body: BodyInit | undefined;
  if (request.method !== "GET" && request.method !== "HEAD") {
    const rawBody = await request.arrayBuffer();
    if (rawBody.byteLength > 0) {
      body = rawBody;
    }
  }

  return fetch(upstreamUrl, {
    method: request.method,
    headers,
    body,
    redirect: "manual"
  });
}

async function handler(request: NextRequest): Promise<Response> {
  try {
    return await proxyRequest(request);
  } catch (error) {
    const message = error instanceof Error ? error.message : "proxy_failed";
    return NextResponse.json({ error: message }, { status: 502 });
  }
}

export const GET = handler;
export const POST = handler;
export const PATCH = handler;
export const PUT = handler;
export const DELETE = handler;
export const OPTIONS = handler;
