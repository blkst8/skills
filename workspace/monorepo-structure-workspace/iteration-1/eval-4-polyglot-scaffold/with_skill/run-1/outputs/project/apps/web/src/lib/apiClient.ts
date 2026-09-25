// User and the request/response shapes come from the generated OpenAPI
// types (@repo/api-schema) — never hand-written, never imported from the
// Go app's source. Regenerate with `make generate` after changing the spec.
import type { components } from "@repo/api-schema";

export type User = components["schemas"]["User"];

const API_URL = import.meta.env.VITE_API_URL ?? "http://localhost:8080";

export const api = {
  async listUsers(): Promise<User[]> {
    const res = await fetch(`${API_URL}/api/v1/users`);
    if (!res.ok) {
      throw new Error(`API error: ${res.status}`);
    }
    return res.json();
  },
};
