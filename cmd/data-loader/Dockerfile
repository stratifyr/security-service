FROM alpine:latest

RUN apk add --no-cache ca-certificates

WORKDIR /src

COPY ./configs /src/configs
COPY ./data-loader /src/data-loader

RUN chmod +x /src/data-loader

ENTRYPOINT ["./data-loader"]