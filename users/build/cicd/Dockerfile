FROM golang:1.26-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
COPY vendor/ vendor/
COPY . .
RUN CGO_ENABLED=0 go build -o /users-server ./cmd/users-server
RUN CGO_ENABLED=0 go build -o /users-worker ./cmd/users-worker

FROM alpine:3.21
RUN apk add --no-cache ca-certificates
COPY --from=builder /users-server /users-server
COPY --from=builder /users-worker /users-worker
EXPOSE 8080
CMD ["/users-server"]
