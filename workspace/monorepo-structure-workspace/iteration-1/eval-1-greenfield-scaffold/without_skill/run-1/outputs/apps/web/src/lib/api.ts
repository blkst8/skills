import type { ApiError, DashboardSummary } from '@dashboard/shared';
import { API_ROUTES } from '@dashboard/shared';

const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL ?? 'http://localhost:3001';

/**
 * Typed client for the dashboard API. Response/request shapes come from
 * @dashboard/shared — the same package the NestJS backend uses, so the
 * contract can never drift between the two apps.
 */
export async function fetchSummary(): Promise<DashboardSummary> {
  const response = await fetch(`${API_BASE_URL}${API_ROUTES.summary}`);

  if (!response.ok) {
    const error = (await response.json()) as ApiError;
    throw new Error(`Failed to load summary (${error.statusCode}): ${error.message}`);
  }

  return (await response.json()) as DashboardSummary;
}
