// API DTOs shared across the frontend.
//
// These hand-mirror the Go structs in apps/api/internal/apitypes/types.go.
// There is no automatic TS<->Go type sharing: keep both sides in sync by hand
// for now, and if the API grows, generate this file from an OpenAPI spec
// (or the Go structs via a codegen step) instead.
//
// Conventions that keep hand-mirroring painless:
//   - Go struct field `json:"snake_case"` tags == the keys here.
//   - Go time.Time serializes as ISO-8601 string == string here.

export interface HealthStatus {
  status: "ok";
  db: boolean;
}

export interface Item {
  id: string;
  title: string;
  created_at: string;
}
