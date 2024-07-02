.PHONY: test

all: lint test

clean:
	-docker stop apicache-dev-redis
	-docker stop apicache-test-redis
	-docker stop apicache-test-memcache

lint:
	gofmt -s -w .
	golangci-lint run --enable-all ./...
	golint ./...

dev: clean
	docker run --rm --name apicache-dev-redis -p 6379:6379 -d redis
	go run cmd/apicache/apicache.go

test: clean
	docker run --rm --name apicache-test-redis -p 6379:6379 -d redis
	docker run --rm --name apicache-test-memcache -p 11211:11211 -d memcached

	go test -race -cover ./...
