FROM golang:1.21 AS builder

WORKDIR /src

COPY go.mod go.sum ./
COPY third_party ./third_party
RUN go mod download

COPY cmd ./cmd
COPY internal ./internal

RUN CGO_ENABLED=0 go build -o /out/scanner ./cmd/scanner

FROM alpine:3.19

RUN apk add --no-cache ca-certificates

COPY --from=builder /out/scanner /usr/local/bin/scanner

WORKDIR /app

ENTRYPOINT ["/usr/local/bin/scanner"]
CMD ["-configPath", "/app/config.yml"]
