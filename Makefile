default: build

.PHONY: build
build:
	go build -o=./cronlock cmd/cronlock/main.go
	go build -o=./cmd/cronlockweb/cronlockweb cmd/cronlockweb/main.go

.PHONY: web
web: 
	cd ./cmd/cronlockweb && ./cronlockweb

.PHONY: clean
clean:
	rm -f ./cronlock
	rm -f ./cmd/cronlockweb/cronlockweb

.PHONY: test
test:
	go test ./...
	