# the module has no git dependency, the alpine builder is enough and much smaller
FROM golang:1.25-alpine AS builder
WORKDIR /go/src/pixelart

# dependencies first: this layer is only rebuilt when go.mod/go.sum change
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o bin/api -v api/main.go

# do not use "scratch" because heroku requires "/bin/sh" as default entrypoint for config vars
# btw, if you don't need config vars you can replace CMD by ENTRYPOINT, it works ;)
FROM alpine:3.22
RUN apk add --no-cache wget \
    && adduser -D -H -u 10001 pixelart
COPY --from=builder /go/src/pixelart/bin/api /app/bin/
# the server never writes to disk, run it as an unprivileged user
USER 10001
EXPOSE 8080
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s \
    CMD wget -qO- "http://127.0.0.1:${PORT:-8080}/health" | grep -q ok || exit 1
CMD ["/app/bin/api"]
