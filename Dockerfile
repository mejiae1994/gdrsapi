FROM golang:1.23.1-alpine AS builder

WORKDIR /build

COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY . .
RUN go build -o gdrsapi ./cmd/api

FROM alpine:latest
WORKDIR /app
COPY --from=builder /build/gdrsapi .

EXPOSE 8081
CMD ["./gdrsapi"]