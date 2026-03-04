FROM golang:1.23-alpine AS builder
WORKDIR /src
COPY mcp/go.mod ./
RUN go mod download
COPY mcp/. .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /bin/mcp ./cmd/server

FROM alpine:3.20
WORKDIR /app
COPY --from=builder /bin/mcp /app/mcp
ENV PORT=8081
EXPOSE 8081
CMD ["/app/mcp"]
