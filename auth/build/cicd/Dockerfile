FROM golang:1.26-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
COPY vendor/ vendor/
COPY . .
RUN CGO_ENABLED=0 go build -o /auth-service ./cmd/auth

FROM alpine:3.21
RUN apk add --no-cache ca-certificates
COPY --from=builder /auth-service /auth-service
EXPOSE 8080
CMD ["/auth-service"]
