.PHONY: build test lint run docker clean

build:
	go build -o bin/wbc ./cmd/wbc

test:
	go test -race ./...

lint:
	go vet ./...
	staticcheck ./...

run: build
	./bin/wbc

docker:
	docker build -t wbc .

clean:
	rm -rf bin/
