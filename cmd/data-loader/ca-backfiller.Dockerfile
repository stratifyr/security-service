FROM golang:1.24.1-alpine AS builder
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -buildvcs=false -o data-loader

FROM alpine:latest
RUN apk add --no-cache ca-certificates tzdata
WORKDIR /src/
COPY --from=builder /src/configs/.env ./configs/.env
COPY --from=builder /src/data-loader ./data-loader
ENTRYPOINT ["sh", "-c", "\
  START_DATE=$(date -d '6 months ago' +%Y-%m-%d) && \
  END_DATE=$(date +%Y-%m-%d) && \
  ./data-loader load security-stats --start-date=$START_DATE --end-date=$END_DATE && \
  ./data-loader load security-metrics\
"]
