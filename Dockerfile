ARG TARGETPLATFORM="linux/amd64"
FROM --platform=${TARGETPLATFORM} golang:1.24-alpine AS builder
ARG TARGETARCH=amd64
ENV GOOS=linux GOARCH=${TARGETARCH} CGO_ENABLED=0 \
    GOPROXY=https://goproxy.cn,direct

WORKDIR /app
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download -x && go mod verify

RUN mkdir -p /app/log && chmod 777 /app/log

COPY . .
COPY ./config.yaml /app/config.yaml
RUN go build -trimpath -ldflags="-w -s" \
    -o /app/main ./cmd/main.go

ARG TARGETPLATFORM
FROM --platform=${TARGETPLATFORM} alpine:3.21
WORKDIR /app

RUN addgroup -S appgroup && adduser -S appuser -G appgroup -u 1001

COPY --from=builder --chown=appuser:appgroup /app/main /app/
COPY --from=builder --chown=appuser:appgroup /app/log /app/log/

RUN apk add --no-cache tzdata && \
    cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime && \
    echo "Asia/Shanghai" > /etc/timezone

USER appuser
EXPOSE 8080
ENTRYPOINT ["/app/main"]