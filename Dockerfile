# use goland:lastest instead of golang:alpine because go git is not available in alpine version
FROM golang:1.18 as builder
WORKDIR /go/src/pixelart
COPY . .
RUN go get -v -d ./...
RUN CGO_ENABLED=0 go build -o bin/api -v api/main.go

FROM scratch
WORKDIR /app
COPY --from=builder /go/src/pixelart/bin/api /app/bin/
EXPOSE 8080
CMD ["./bin/api"]
