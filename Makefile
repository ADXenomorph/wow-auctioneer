.PHONY: test generate-mocks
test:
	go test -v -race -timeout 30s -cover ./...

generate-mocks:
	find . -name '*_minimock.go' -delete
	go generate ./...
	go mod tidy -compat=1.17

dep:
	go mod download
	go mod tidy
	go mod vendor

build: dep
	go build -mod vendor -o wow-auctioneer .
