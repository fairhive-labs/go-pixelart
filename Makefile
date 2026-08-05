.PHONY: run build cli clean test race bench vet fmt

run: build
	./bin/api
build: clean
	go build -trimpath -o bin/api -v api/main.go
cli: clean
	go build -trimpath -o bin/pixelart -v ./cmd
clean:
	rm -rf ./bin
test:
	go clean -testcache
	go test ./...
race:
	go test -race ./...
bench:
	go test ./internal/... -run XXX -bench . -benchmem
vet:
	go vet ./...
fmt:
	gofmt -l -w .
