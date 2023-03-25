# use goland:lastest instead of golang:alpine because go git is not available in alpine version
FROM golang:1.20 as builder
WORKDIR /go/src/pixelart
COPY . .
RUN go mod tidy
RUN CGO_ENABLED=0 go build -o bin/api -v api/main.go

# do not use "scratch" because heroku requires "/bin/sh" as default entrypoint for config vars
# btw, if you don't need config vars you can replace CMD by ENTRYPOINT, it works ;)
FROM alpine
COPY --from=builder /go/src/pixelart/bin/api /app/bin/
EXPOSE 8080
CMD ["/app/bin/api"]
