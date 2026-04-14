FROM golang:1.25-alpine AS builder

RUN apk add --no-cache git

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -o j-support ./cmd/j-support

FROM alpine:3.23

WORKDIR /build

COPY --from=builder /build/j-support /build/j-support

COPY --from=builder /build/migrations /build/migrations

COPY .env ./

EXPOSE 7777

ENTRYPOINT ["/build/j-support"]