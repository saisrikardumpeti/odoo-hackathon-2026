FROM node:22-alpine AS base
WORKDIR /app
# Enable corepack for built-in pnpm support
RUN corepack enable pnpm

# --- Development Stage (Targeted by docker-compose) ---
FROM base AS dev
COPY package.json pnpm-lock.yaml ./
RUN pnpm install
# Note: Source code is bind-mounted at runtime
EXPOSE 5173
CMD ["pnpm", "run", "dev", "--host"]
