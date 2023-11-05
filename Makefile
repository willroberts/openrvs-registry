.PHONY: lint
lint:
	golint ./...
	go vet ./...

.PHONY: test
test:
	go test ./...

.PHONY: test_integ
test_integ:
	go test -tags=integration ./...

.PHONY: coverage
coverage:
	go test -v -tags=integration -coverprofile=cover.out
	go tool cover -func=cover.out