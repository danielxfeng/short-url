FROM golang:1.25.5-alpine AS builder
WORKDIR /app
RUN apk add --no-cache build-base
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=1 GOOS=linux go build -o /app/bin/quick-agent-factory ./cmd/server

FROM alpine:3.20
RUN apk add --no-cache ca-certificates sqlite-libs \
    && addgroup -S app \
    && adduser -S app -G app

WORKDIR /app
COPY --from=builder /app/bin/quick-agent-factory /usr/local/bin/quick-agent-factory

RUN mkdir -p /app/data \
    && chown -R app:app /app

ENV APP_MODE=production

EXPOSE 8080
USER app
ENTRYPOINT ["quick-agent-factory"]