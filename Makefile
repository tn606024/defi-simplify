.PHONY: help test check require-base-rpc anvil-base test-integration update-aave-manifest update-aave-manifests

ANVIL_BIN ?= anvil
ANVIL_HOST ?= 127.0.0.1
ANVIL_PORT ?= 8545
BASE_CHAIN_ID ?= 8453
ANVIL_HARDFORK ?= prague

help:
	@printf "Available targets:\n"
	@printf "  make test              Run unit tests\n"
	@printf "  make check             Run unit tests and whitespace checks\n"
	@printf "  make anvil-base        Start an Anvil fork of Base mainnet (requires BASE_RPC_URL)\n"
	@printf "  make test-integration  Run integration tests (requires BASE_RPC_URL)\n"
	@printf "  make update-aave-manifests Regenerate reviewed Base Aave manifests\n"

test:
	go test ./...

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
	$(ANVIL_BIN) --fork-url $(BASE_RPC_URL) --chain-id $(BASE_CHAIN_ID) --hardfork $(ANVIL_HARDFORK) --host $(ANVIL_HOST) --port $(ANVIL_PORT)

test-integration: require-base-rpc
	BASE_RPC_URL=$(BASE_RPC_URL) go test -count=1 -tags=integration ./integration/...

update-aave-manifest: update-aave-manifests

update-aave-manifests:
	npm ci --ignore-scripts --legacy-peer-deps --prefix tools/aave-address-book
	go run ./cmd/update-aave-manifest
