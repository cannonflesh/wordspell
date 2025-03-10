.PHONY: dep-install test

dep-install:
	go mod download

test:
	go test -cover ./...