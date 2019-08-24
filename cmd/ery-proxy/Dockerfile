FROM golang:1.12-alpine3.10 as builder

ENV GO111MODULE on
WORKDIR /go/github.com/srvc/ery

RUN apk add --update --no-cache git

COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . .
RUN go build -o /usr/local/bin/ery-proxy ./cmd/ery-proxy

FROM alpine:3.10

COPY --from=builder /usr/local/bin/ery-proxy /usr/local/bin/
