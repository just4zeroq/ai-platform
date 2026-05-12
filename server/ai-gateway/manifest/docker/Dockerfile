FROM golang:1.25-alpine AS builder
WORKDIR /build
COPY api/ api/
COPY ai-gateway/ ai-gateway/
WORKDIR /build/ai-gateway
RUN go build -o /app/main .

FROM alpine:3.19
RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app
COPY --from=builder /app/main .
COPY --from=builder /build/ai-gateway/manifest/ manifest/
COPY --from=builder /build/ai-gateway/resource/ resource/
COPY docker/entrypoint.sh /app/entrypoint.sh
RUN chmod +x /app/entrypoint.sh
ENTRYPOINT ["/app/entrypoint.sh"]
EXPOSE 8081
