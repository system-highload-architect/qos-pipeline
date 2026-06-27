FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY . .
WORKDIR /app/services/gateway
RUN go mod download
WORKDIR /app
RUN CGO_ENABLED=0 GOOS=linux go build -o /gateway ./services/gateway/cmd

FROM alpine:3.19
RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=builder /gateway .
COPY services/gateway/configs/ ./configs/
EXPOSE 8080
ENTRYPOINT ["./gateway"]