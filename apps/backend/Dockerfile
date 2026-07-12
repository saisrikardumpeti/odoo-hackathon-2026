FROM golang:1.25-alpine AS base
WORKDIR /app

# --- Development Stage (Targeted by docker-compose) ---
FROM base AS dev
# Install air for hot-reloading
RUN go install github.com/air-verse/air@latest
COPY go.mod go.sum ./
RUN go mod download
# Note: Source code is bind-mounted at runtime
EXPOSE 8000
CMD ["air"]
