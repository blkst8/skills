# syntax=docker/dockerfile:1
# Build from the repo root: docker build -f apps/api/Dockerfile .

FROM node:20-alpine AS build
WORKDIR /repo
RUN corepack enable
COPY package.json pnpm-workspace.yaml pnpm-lock.yaml ./
COPY apps/api/package.json apps/api/
COPY packages/config/package.json packages/config/
COPY packages/shared/package.json packages/shared/
COPY packages/types/package.json packages/types/
RUN pnpm install --frozen-lockfile
COPY . .
RUN pnpm --filter @dashboard/api... build \
    && pnpm --filter @dashboard/api --prod deploy /deployed

FROM node:20-alpine AS runtime
WORKDIR /app
ENV NODE_ENV=production
COPY --from=build /deployed ./
EXPOSE 4000
CMD ["node", "dist/main.js"]
