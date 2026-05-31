# syntax=docker/dockerfile:1

# ---------- 1. 构建前端 ----------
FROM node:22-alpine AS frontend
WORKDIR /build/frontend
COPY frontend/package.json frontend/package-lock.json ./
RUN npm ci
COPY frontend/ ./
RUN npm run build

# ---------- 2. 构建后端 ----------
FROM golang:1.25-alpine AS backend
WORKDIR /build/backend
RUN apk add --no-cache git
COPY backend/go.mod backend/go.sum ./
RUN go mod download
COPY backend/ ./
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /server .

# ---------- 3. 运行镜像 ----------
FROM alpine:3.21
RUN apk add --no-cache ca-certificates tzdata \
    && adduser -D -h /app appuser
WORKDIR /app

COPY --from=backend /server ./server
COPY --from=frontend /build/frontend/dist ./static

ENV GIN_MODE=release \
    PORT=8080 \
    DATA_DIR=/app/data \
    TZ=Asia/Shanghai

RUN mkdir -p /app/data/uploads \
    && chown -R appuser:appuser /app

USER appuser
EXPOSE 8080
VOLUME ["/app/data"]

HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
    CMD wget -qO- http://127.0.0.1:8080/api/dishes >/dev/null || exit 1

CMD ["./server"]
