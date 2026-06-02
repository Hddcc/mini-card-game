# syntax=docker/dockerfile:1

FROM golang:1.25-alpine AS builder

WORKDIR /src

RUN apk add --no-cache ca-certificates git tzdata

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/mini-card-game ./cmd/server

FROM alpine:3.22

WORKDIR /app

RUN apk add --no-cache ca-certificates tzdata wget \
    && addgroup -S app \
    && adduser -S app -G app

ENV APP_ENV=production \
    HTTP_ADDR=:5290 \
    FRONTEND_DIST=/app/frontend/stitch \
    TZ=Asia/Shanghai

COPY --from=builder /out/mini-card-game /app/mini-card-game
COPY --from=builder /src/frontend/stitch /app/frontend/stitch

USER app

EXPOSE 5290

HEALTHCHECK --interval=30s --timeout=3s --start-period=30s --retries=3 \
    CMD wget -qO- http://127.0.0.1:5290/health >/dev/null || exit 1

ENTRYPOINT ["/app/mini-card-game"]
