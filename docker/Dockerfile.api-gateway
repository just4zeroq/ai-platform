FROM golang:1.25-alpine AS builder
WORKDIR /build
COPY api/ api/
COPY api-gateway/ api-gateway/
WORKDIR /build/api-gateway
RUN go build -o /app/main .

FROM alpine:3.19
RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app
COPY --from=builder /app/main .
COPY --from=builder /build/api-gateway/manifest/ manifest/
COPY --from=builder /build/api-gateway/resource/ resource/
COPY docker/entrypoint.sh /app/entrypoint.sh
RUN chmod +x /app/entrypoint.sh
ENTRYPOINT ["/app/entrypoint.sh"]
EXPOSE 8080
