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
# run the redis tests first to ensure the redis server is running
	go test -v ./common/redis
# run the rest of the tests
	go test -v ./common/cron
	