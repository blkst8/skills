/**
 * @dashboard/shared — the single source of truth for the API contract.
 *
 * The backend (`apps/api`) imports these types for its controllers/DTOs, and
 * the frontend (`apps/web`) imports the same types to type its API calls.
 * Change the contract here, and both apps see it on the next build.
 */

/* ---------------------------------- Domain --------------------------------- */

export type WidgetType = 'metric' | 'chart' | 'table';

export interface Widget {
  id: string;
  title: string;
  type: WidgetType;
  /** Arbitrary, widget-type-specific payload (chart series, metric value, ...) */
  data: Record<string, unknown>;
  createdAt: string; // ISO 8601
  updatedAt: string; // ISO 8601
}

export interface DashboardSummary {
  totalWidgets: number;
  lastUpdated: string; // ISO 8601
}

/* ------------------------------ API contract ------------------------------- */

/** Request payload for POST /widgets */
export interface CreateWidgetDto {
  title: string;
  type: WidgetType;
  data?: Record<string, unknown>;
}

/** Request payload for PATCH /widgets/:id */
export interface UpdateWidgetDto {
  title?: string;
  data?: Record<string, unknown>;
}

/** Error envelope returned by the API on failures */
export interface ApiError {
  statusCode: number;
  message: string;
  error?: string;
}

/** Route constants so the frontend never hard-codes paths */
export const API_ROUTES = {
  widgets: '/widgets',
  widgetById: (id: string) => `/widgets/${id}`,
  summary: '/summary',
} as const;
