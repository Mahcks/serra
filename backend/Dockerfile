# Choose Go version
ARG GOLANG_TAG=1.24-alpine

FROM golang:${GOLANG_TAG} AS builder

# Environment variables for bootstrap config
ENV REST_ADDRESS=""
ENV REST_PORT=""
ENV SQLITE_PATH=""
ENV CREDENTIALS_JWT_SECRET=""

ARG VERSION=""
ARG COMMIT=""

# Install required tools for CGO + SQLite
RUN apk add --no-cache ca-certificates git build-base sqlite-dev

RUN apk add --no-cache ca-certificates git

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=1 GOOS=linux go build \
    -o serra-server \
    -ldflags="-s -w -X 'main.Version=${VERSION}' -X 'main.CommitHash=${COMMIT}'" \
    ./cmd/app/main.go

# Add Atlas for migrations
FROM alpine:latest AS atlas
RUN apk add --no-cache curl
RUN curl -sSf https://atlasgo.sh | sh

# Use distroless instead of scratch
FROM alpine:latest
RUN apk add --no-cache ca-certificates sqlite-libs

WORKDIR /app

COPY --from=builder /app/serra-server /app/serra-server
COPY --from=atlas /usr/local/bin/atlas /usr/local/bin/atlas
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /app/migrations /app/app/migrations

ENV SQLITE_PATH="/app/data.db"

COPY config/config.yaml /home/nonroot/config/config.yaml

ENTRYPOINT sh -c "\
    atlas migrate apply --url sqlite3://$SQLITE_PATH --dir file://app/migrations && \
    /app/serra-server"
