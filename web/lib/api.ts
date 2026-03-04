import { MemberConnection, Vacation } from "./types";

const apiBaseUrl = process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8080";

type ConnectPayload = {
  displayName: string;
  goal: string;
  colorHex: string;
  adminSecret?: string;
};

export async function connectMember(payload: ConnectPayload): Promise<MemberConnection> {
  const response = await fetch(`${apiBaseUrl}/v1/connect`, {
    method: "POST",
    headers: {
      "Content-Type": "application/json"
    },
    body: JSON.stringify(payload),
    cache: "no-store"
  });

  if (!response.ok) {
    throw new Error(`Connect failed: ${response.status}`);
  }

  return response.json();
}

export async function getVacations(year: number): Promise<Vacation[]> {
  const response = await fetch(`${apiBaseUrl}/v1/vacations?year=${year}`, {
    next: { revalidate: 0 }
  });

  if (!response.ok) {
    throw new Error(`Vacations request failed: ${response.status}`);
  }

  return response.json();
}
