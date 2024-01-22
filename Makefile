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
	mkdir -p bin
	$(eval VERSION = $(shell git describe --tags --abbrev=0))
	GOOS=darwin GOARCH=arm64 go build -o bin/registry-${VERSION}-darwin-arm64 cmd/registry/main.go
	GOOS=linux GOARCH=amd64 go build -o bin/registry-${VERSION}-linux-amd64 cmd/registry/main.go
	GOOS=windows GOARCH=amd64 go build -o bin/registry-${VERSION}-windows-amd64.exe cmd/registry/main.go

.PHONY: run
run:
	go run cmd/registry/main.go
