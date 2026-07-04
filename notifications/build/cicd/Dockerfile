FROM golang:1.26-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
COPY vendor/ vendor/
COPY . .
RUN CGO_ENABLED=0 go build -o /notifications-http ./cmd/notifications-http
RUN CGO_ENABLED=0 go build -o /notifications-worker ./cmd/notifications-worker

FROM alpine:3.21
RUN apk add --no-cache ca-certificates
COPY --from=builder /notifications-http /notifications-http
COPY --from=builder /notifications-worker /notifications-worker
EXPOSE 8080
CMD ["/notifications-http"]
