FROM golang:1.26-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
COPY vendor/ vendor/
COPY . .
RUN CGO_ENABLED=0 go build -o /tasks-http ./cmd/tasks-http

FROM alpine:3.21
RUN apk add --no-cache ca-certificates
COPY --from=builder /tasks-http /tasks-http
EXPOSE 8080
CMD ["/tasks-http"]
