# use goland:lastest instead of golang:alpine because go git is not available in alpine version
FROM golang:1.18 as builder
WORKDIR /go/src/pixelart
COPY . .
RUN go mod tidy
RUN CGO_ENABLED=0 go build -o bin/pixelart -v api/main.go

# do not use scratch because heroku requires /bin/sh as default entrypoint for config vars
FROM scratch
COPY --from=builder /go/src/pixelart/bin/pixelart /app/bin/
EXPOSE 8080
ENTRYPOINT ["/app/bin/pixelart"]
