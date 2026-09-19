// Request/response envelopes and cross-cutting API shapes.

export interface ApiEnvelope<T> {
  data: T;
}

export interface ApiError {
  statusCode: number;
  message: string;
  error?: string;
}

export interface HealthCheckDto {
  status: "ok";
  uptime: number;
  timestamp: string;
}
