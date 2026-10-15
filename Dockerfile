############################
# 1) Build stage
############################
FROM --platform=$BUILDPLATFORM golang:1.25.3-alpine AS builder

WORKDIR /src

RUN apk add --no-cache ca-certificates tzdata git

ARG TARGETOS=linux
ARG TARGETARCH
ENV CGO_ENABLED=0 \
    GOOS=${TARGETOS} \
    GOARCH=${TARGETARCH} \
    GOPROXY=https://proxy.golang.org,direct

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN mkdir -p pkg && go build -trimpath -ldflags="-s -w" -o pkg/jh_app_service .

############################
# 2) Runtime stage
############################
FROM alpine:3.20

WORKDIR /app

RUN apk add --no-cache ca-certificates tzdata \
    && addgroup -S app && adduser -S -G app app

COPY --from=builder /src/pkg/jh_app_service /app/jh_app_service
COPY --from=builder /src/manifest/config /app/manifest/config

EXPOSE 8001

USER app
CMD ["/app/jh_app_service"]
