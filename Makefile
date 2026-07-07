.PHONY: help test check require-base-rpc anvil-base test-integration

GOCACHE ?= /private/tmp/defi-simplify-gocache
ANVIL_BIN ?= anvil
ANVIL_HOST ?= 127.0.0.1
ANVIL_PORT ?= 8545
BASE_CHAIN_ID ?= 8453

help:
	@printf "Available targets:\n"
	@printf "  make test              Run unit tests\n"
	@printf "  make check             Run unit tests and whitespace checks\n"
	@printf "  make anvil-base        Start an Anvil fork of Base mainnet (requires BASE_RPC_URL)\n"
	@printf "  make test-integration  Run integration tests (requires BASE_RPC_URL)\n"

test:
	GOCACHE=$(GOCACHE) go test ./...

check: test
	git diff --check

require-base-rpc:
	@if [ -z "$(BASE_RPC_URL)" ]; then \
		echo "BASE_RPC_URL is required"; \
		echo "Example: BASE_RPC_URL=https://mainnet.base.org make anvil-base"; \
		echo "Example: BASE_RPC_URL=http://127.0.0.1:8545 make test-integration"; \
		exit 1; \
	fi

anvil-base: require-base-rpc
	$(ANVIL_BIN) --fork-url $(BASE_RPC_URL) --chain-id $(BASE_CHAIN_ID) --host $(ANVIL_HOST) --port $(ANVIL_PORT)

test-integration: require-base-rpc
	GOCACHE=$(GOCACHE) BASE_RPC_URL=$(BASE_RPC_URL) go test -tags=integration ./integration/...
