FROM golang:1.26.4-alpine3.23
WORKDIR /app
RUN go install github.com/air-verse/air@latest
COPY go.mod go.sum ./
COPY vendor/ ./vendor/
COPY . .
ENV GOFLAGS="-mod=vendor"
CMD ["air", "-c", ".air.toml"]
