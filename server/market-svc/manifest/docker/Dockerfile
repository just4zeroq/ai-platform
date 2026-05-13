FROM golang:1.26-alpine AS builder
WORKDIR /build
COPY api/ api/
COPY market-svc/ market-svc/
WORKDIR /build/market-svc
RUN go build -o /app/main .

FROM alpine:latest
RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app
COPY --from=builder /app/main .
COPY --from=builder /build/market-svc/manifest/ manifest/
COPY --from=builder /build/market-svc/resource/ resource/
COPY docker/entrypoint.sh /app/entrypoint.sh
RUN chmod +x /app/entrypoint.sh
ENTRYPOINT ["/app/entrypoint.sh"]
EXPOSE 8102
