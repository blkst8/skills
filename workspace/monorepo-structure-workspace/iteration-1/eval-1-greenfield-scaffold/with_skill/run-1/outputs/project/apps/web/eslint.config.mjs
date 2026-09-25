import { FlatCompat } from "@eslint/eslintrc";
import base from "@repo/config/eslint";

const compat = new FlatCompat({ baseDirectory: import.meta.dirname });

export default [
  ...base,
  ...compat.extends("next/core-web-vitals", "next/typescript"),
];
