FROM golang:1.20.4-alpine3.17 AS builder

WORKDIR /app
COPY src ./

RUN CGO_ENABLED=0 GOOS=linux go build -o /basiccoin ./fullnode

FROM alpine:3.18.0 AS main

COPY --from=builder /basiccoin /

EXPOSE 80

CMD ["/basiccoin"]

FROM golang:1.20.4-alpine3.17 AS cli-builder

WORKDIR /app
COPY src ./

RUN CGO_ENABLED=0 GOOS=linux go build -o /basiccoin-cli ./cli

FROM alpine:3.18.0 AS cli

COPY --from=cli-builder /basiccoin-cli /

ENTRYPOINT /basiccoin-cli help
