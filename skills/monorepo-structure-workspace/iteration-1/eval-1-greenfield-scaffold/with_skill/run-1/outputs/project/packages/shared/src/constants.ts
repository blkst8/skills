// Route paths served by apps/api — the single source of truth for both
// apps, so a renamed route changes here and nowhere else.
export const API_ROUTES = {
  health: "/health",
  users: "/users",
} as const;
