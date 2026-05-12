FROM golang:1.25-alpine AS builder
WORKDIR /build
COPY api/ api/
COPY user-svc/ user-svc/
WORKDIR /build/user-svc
RUN go build -o /app/main .

FROM alpine:3.19
RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app
COPY --from=builder /app/main .
COPY --from=builder /build/user-svc/manifest/ manifest/
COPY --from=builder /build/user-svc/resource/ resource/
COPY docker/entrypoint.sh /app/entrypoint.sh
RUN chmod +x /app/entrypoint.sh
ENTRYPOINT ["/app/entrypoint.sh"]
EXPOSE 8100
