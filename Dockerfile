FROM golang:1.26.2-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o /shelfshare .

FROM scratch

COPY --from=builder /shelfshare /shelfshare

EXPOSE 8080

ENTRYPOINT ["/shelfshare"]
