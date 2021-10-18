.PHONY: build run clean build_client build_server run_client run_server

build_client:
	@go build -o ./bin/client ./cmd/client

build_server:
	@go build -o ./bin/server ./cmd/server

build:
	@$(MAKE) build_client
	@$(MAKE) build_server

run_client: build_client
	@./bin/client

run_server: build_server
	@./bin/server

run: build
	@$(MAKE) run_client
	@$(MAKE) run_server

clean:
	@rm -rf ./bin/*

