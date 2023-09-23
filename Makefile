run: clean build
	./bin/api
build: clean
	go build -o bin/api -v api/main.go
clean:
	rm -rf ./bin
test:
	go clean -testcache
	go test ./... 