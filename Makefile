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
	go test -v -tags=integration -coverprofile=cover.out
	go tool cover -func=cover.out