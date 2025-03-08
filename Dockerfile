FROM golang:1.24-alpine AS builder
RUN apk add --no-cache gcc musl-dev
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=1 GOOS=linux go build -a -o url-shortener ./cmd

FROM alpine:latest
RUN apk add --no-cache libc6-compat ca-certificates sqlite
WORKDIR /app
COPY --from=builder /app/url-shortener .
COPY --from=builder /app/templates ./templates
COPY --from=builder /app/static ./static
RUN mkdir -p /data
EXPOSE 8080
ENV PORT=8080
ENV DB_PATH=/data/data.db
ENV BASE_URL=http://localhost:8080
VOLUME ["/data"]
CMD ./url-shortener --port ${PORT} --db ${DB_PATH} --base-url ${BASE_URL} 