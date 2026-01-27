run:
	@go run ./cmd/api

test:
	@go test -v ./...

test-integration:
	@go test -v -tags=integration ./...