FROM golang:1.23-alpine AS builder
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /bin/api ./cmd/server

FROM alpine:3.20
WORKDIR /app
COPY --from=builder /bin/api /app/api
ENV PORT=8080
EXPOSE 8080
CMD ["/app/api"]
