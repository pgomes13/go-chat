.PHONY: run build test deploy clean

run:
	go run ./cmd/server

build:
	go build -o bin/server ./cmd/server

test:
	go test ./... -v -timeout 30s

deploy:
	./scripts/deploy/main.sh

clean:
	rm -rf bin/
