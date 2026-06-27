FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY . .
WORKDIR /app/services/ingester
RUN go mod download
WORKDIR /app
RUN CGO_ENABLED=0 GOOS=linux go build -o /service ./services/ingester/cmd

FROM alpine:3.19
RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=builder /service .
COPY services/ingester/configs/ ./configs/
EXPOSE 9001
ENTRYPOINT ["./service"]