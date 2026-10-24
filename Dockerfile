FROM golang:1.26.0-alpine AS builder
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -buildvcs=false -o security-service

FROM alpine:3.24
RUN apk add --no-cache ca-certificates && addgroup -S app && adduser -S -G app app
WORKDIR /src/
COPY --from=builder --chown=app:app /src/configs/.env ./configs/.env
COPY --from=builder --chown=app:app /src/security-service ./security-service

USER app
EXPOSE 8000 9000
HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
    CMD wget -q -O /dev/null http://127.0.0.1:8000/.well-known/health || exit 1
ENTRYPOINT ["./security-service"]
