build: clean
	go build -o bin/api -v api/main.go
run: clean build
	./bin/api
clean:
	rm -rf ./bin
test:
	go clean -testcache
	go test -v ./... 