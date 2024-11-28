FROM golang:1.22 as builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o rate-limiter

FROM scratch
WORKDIR /app
COPY --from=builder /app/rate-limiter .
COPY --from=builder /app/.env .
ENTRYPOINT ["./rate-limiter"]