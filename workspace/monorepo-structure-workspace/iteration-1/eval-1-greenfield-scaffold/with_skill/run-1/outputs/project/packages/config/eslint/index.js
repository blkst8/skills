import eslint from "@eslint/js";
import tseslint from "typescript-eslint";

// Shared flat-config base. Apps import this array and append
// framework-specific rules (Next.js, NestJS) on top of it.
export default tseslint.config(
  {
    ignores: [
      "**/node_modules/**",
      "**/dist/**",
      "**/.next/**",
      "**/.turbo/**",
      "**/coverage/**",
    ],
  },
  eslint.configs.recommended,
  ...tseslint.configs.recommended,
  {
    rules: {
      // Keeps type-only sharing cheap: contract imports compile away.
      "@typescript-eslint/consistent-type-imports": [
        "warn",
        { fixStyle: "inline-type-imports" },
      ],
    },
  },
);
