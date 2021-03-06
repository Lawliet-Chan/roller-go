# Build roller in a stock Go builder container
FROM golang:1.17-alpine as builder

ENV GOPROXY https://goproxy.io,direct

ADD . /roller-go
RUN apk add --no-cache gcc musl-dev linux-headers git ca-certificates \
    && cd /roller-go/cmd/roller/ && go build -v -p 4

# Pull roller into a second stage deploy alpine container
FROM alpine:latest

COPY --from=builder /roller-go/cmd/roller/roller /bin/

ENTRYPOINT ["roller"]
