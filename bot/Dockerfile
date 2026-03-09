FROM golang:1.23-alpine AS builder
WORKDIR /src
COPY bot/go.mod ./
RUN go mod download
COPY bot/. .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /bin/bot ./cmd/server

FROM alpine:3.20
WORKDIR /app
COPY --from=builder /bin/bot /app/bot
ENV PORT=8082
EXPOSE 8082
CMD ["/app/bot"]
