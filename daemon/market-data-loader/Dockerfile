FROM golang:1.26.0-alpine AS builder
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -buildvcs=false -o market-data-loader

FROM alpine:latest
RUN apk add --no-cache ca-certificates
WORKDIR /src/
COPY --from=builder /src/configs/.env ./configs/.env
COPY --from=builder /src/market-data-loader ./market-data-loader
ENTRYPOINT ["./market-data-loader"]
