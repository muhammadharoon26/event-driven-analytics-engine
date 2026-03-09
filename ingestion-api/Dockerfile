FROM golang:1.23.5 AS builder

WORKDIR /app

COPY go.mod go.sum* ./
RUN go mod download

COPY . .

# Build the application statically without CGO for alpine compatibility
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o ingestion-api main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates libc6-compat

WORKDIR /root/
COPY --from=builder /app/ingestion-api .

EXPOSE 8080
CMD ["./ingestion-api"]
