FROM golang:1.20.5-alpine3.18 AS builder

WORKDIR /basiccoin
COPY cmd ./cmd
COPY internal ./internal
COPY pkg ./pkg
COPY go.mod ./

RUN echo $(ls -l ./)

RUN CGO_ENABLED=0 GOOS=linux go build -o /bcnode ./cmd/bcnode

FROM alpine:3.18.2 AS main

COPY --from=builder /bcnode /

EXPOSE 80

CMD ["/bcnode"]
