.PHONY: garnet

run:
	@go build -o ./build/indexer ./cmd/indexer && ./build/indexer http://localhost:8545

lint:
	golangci-lint run --fix --out-format=line-number --issues-exit-code=0 --config .golangci.yml --color always ./...

