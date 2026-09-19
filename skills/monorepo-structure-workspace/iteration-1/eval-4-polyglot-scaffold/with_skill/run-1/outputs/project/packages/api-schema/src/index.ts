// Re-exports the types generated from openapi.yaml.
// After changing the spec, run `make generate` (or
// `pnpm --filter api-schema generate`) — never hand-edit the generated file.
export type { components, paths } from "./generated/schema-types";
