FROM cgr.dev/chainguard/go AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN go build -o service .

FROM cgr.dev/chainguard/busybox:latest AS runner

WORKDIR /usr/bin

COPY --from=builder /app/service ./
COPY --from=builder /app/docs ./docs
COPY --from=builder /app/migrations ./migrations

EXPOSE 3000

ENTRYPOINT ["./service", "serve"]
