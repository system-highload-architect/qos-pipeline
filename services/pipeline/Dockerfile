FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY . .
WORKDIR /app/services/pipeline
RUN go mod download
WORKDIR /app
RUN CGO_ENABLED=0 GOOS=linux go build -o /service ./services/pipeline/cmd

FROM alpine:3.19
RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=builder /service .
COPY services/pipeline/configs/ ./configs/
EXPOSE 9000
ENTRYPOINT ["./service"]