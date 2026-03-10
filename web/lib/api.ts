import { DayOffOverride, Vacation } from "./types";

const apiBaseUrl =
  process.env.NEXT_PUBLIC_API_URL ??
  (process.env.NODE_ENV === "development" ? "http://localhost:8080" : "");

function getApiBaseUrl(): string {
  if (apiBaseUrl) {
    return apiBaseUrl;
  }

  throw new Error("NEXT_PUBLIC_API_URL is not set");
}

export async function getVacations(year: number): Promise<Vacation[]> {
  const response = await fetch(`${getApiBaseUrl()}/v1/vacations?year=${year}`, {
    next: { revalidate: 0 }
  });

  if (!response.ok) {
    throw new Error(`Vacations request failed: ${response.status}`);
  }

  return response.json();
}

export async function getDayOffOverrides(year: number): Promise<DayOffOverride[]> {
  const response = await fetch(`${getApiBaseUrl()}/v1/dayoffs?year=${year}`, {
    next: { revalidate: 0 }
  });

  if (!response.ok) {
    throw new Error(`Day offs request failed: ${response.status}`);
  }

  return response.json();
}
