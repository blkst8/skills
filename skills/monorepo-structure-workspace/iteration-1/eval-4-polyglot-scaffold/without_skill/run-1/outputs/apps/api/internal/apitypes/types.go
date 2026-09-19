package apitypes

import "time"

// DTOs serialized to the frontend.
//
// These hand-mirror the TypeScript interfaces in packages/types/src/index.ts.
// There is no automatic TS<->Go type sharing: keep both sides in sync by hand
// for now, and if the API grows, generate the TS types from an OpenAPI spec
// (or from these structs via a codegen step) instead.
//
// The json tags below MUST match the TypeScript field names exactly.

type HealthStatus struct {
	Status string `json:"status"`
	DB     bool   `json:"db"`
}

type Item struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	CreatedAt time.Time `json:"created_at"`
}
