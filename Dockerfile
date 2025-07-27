FROM golang:1.24.1-alpine AS builder
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -buildvcs=false -o security-service

FROM alpine:latest
RUN apk add --no-cache ca-certificates
WORKDIR /src/
COPY --from=builder /src/configs/.env ./configs/.env
COPY --from=builder /src/security-service ./security-service
ENTRYPOINT ["./security-service"]
