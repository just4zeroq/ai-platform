FROM golang:1.26-alpine AS builder
WORKDIR /build
COPY api/ api/
COPY asset-svc/ asset-svc/
WORKDIR /build/asset-svc
RUN go build -o /app/main .

FROM alpine:latest
RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app
COPY --from=builder /app/main .
COPY --from=builder /build/asset-svc/manifest/ manifest/
COPY --from=builder /build/asset-svc/resource/ resource/
COPY docker/entrypoint.sh /app/entrypoint.sh
RUN chmod +x /app/entrypoint.sh
ENTRYPOINT ["/app/entrypoint.sh"]
EXPOSE 8101
