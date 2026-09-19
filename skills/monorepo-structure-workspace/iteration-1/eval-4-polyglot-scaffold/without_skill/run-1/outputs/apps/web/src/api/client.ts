import type { HealthStatus } from "@sideproject/types";

const BASE = "/api";

export async function fetchHealth(): Promise<HealthStatus> {
  const res = await fetch(`${BASE}/healthz`);
  if (!res.ok) {
    throw new Error(`health check failed: ${res.status}`);
  }
  return res.json();
}
