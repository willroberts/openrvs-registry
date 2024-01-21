.PHONY: lint
lint:
	golint ./...
	go vet ./...

.PHONY: test
test:
	go test -v ./...

.PHONY: test_integ
test_integ:
	go test -v -tags=integration ./...

.PHONY: coverage
coverage:
	go test -v -tags=integration -coverprofile=cover.out ./...
	go tool cover -html=cover.out

.PHONY: build
build:
	mkdir bin
	go build -o bin/registry cmd/registry/main.go

.PHONY: run
run:
	go run cmd/registry/main.go
