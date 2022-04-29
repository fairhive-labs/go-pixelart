# use goland:lastest instead of golang:alpine because go git is not available in alpine version
FROM golang:1.18 as builder
WORKDIR /go/src/pixelart
COPY . .
RUN go mod tidy
RUN CGO_ENABLED=0 go build -o bin/pixelart -v api/main.go

FROM alpine
COPY --from=builder /go/src/pixelart/bin/pixelart /app/bin/
EXPOSE 8080
CMD ["/app/bin/pixelart"]
