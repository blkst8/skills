import base from "@repo/config/eslint";

export default [
  ...base,
  {
    rules: {
      // Nest DI and decorator interop occasionally require `any` at seams;
      // warn there instead of failing the build.
      "@typescript-eslint/no-explicit-any": "warn",
    },
  },
];
