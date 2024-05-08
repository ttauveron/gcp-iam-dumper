FROM golang:1.21.6-alpine as builder

WORKDIR /app
COPY go.* ./
RUN go mod download
COPY . .

ENV CGO_ENABLED=1
ENV GOOS=linux

RUN apk add --no-cache \
    gcc \
    musl-dev

RUN go build -o main ./cmd/

FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=builder /app/main ./gcp-iam-dumper

CMD ["./gcp-iam-dumper"]
