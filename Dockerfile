# use goland:lastest instead of golang:alpine because go git is not available in alpine version
FROM golang as builder
WORKDIR /go/src/pixelart
COPY . .
RUN go get -v -d ./...
RUN go build -o bin/api -v api/main.go

FROM alpine
WORKDIR /app
COPY --from=builder /go/src/pixelart/bin/api /app/bin/
RUN apk add --no-cache bash
CMD ["./bin/api"]
