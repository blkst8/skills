/** @type {import('next').NextConfig} */
const nextConfig = {
  output: "standalone",
  // Internal packages ship raw TypeScript — let Next compile them.
  transpilePackages: ["@repo/types", "@repo/shared"],
};

export default nextConfig;
