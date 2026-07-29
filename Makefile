.PHONY: help test check require-base-rpc require-contracts-dir anvil-base test-integration update-aave-manifest update-aave-manifests update-contract-artifacts generate-contract-bindings verify-contract-artifacts generate-weth-binding verify-weth-binding

ANVIL_BIN ?= anvil
ANVIL_HOST ?= 127.0.0.1
ANVIL_PORT ?= 8545
BASE_CHAIN_ID ?= 8453
ANVIL_HARDFORK ?= prague
CONTRACTS_DIR ?= $(DEFI_SIMPLIFY_CONTRACTS_DIR)
CONTRACT_ARTIFACT_DIR := client/account/defisimplify7702
CONTRACT_BINDING_DIR := $(CONTRACT_ARTIFACT_DIR)/bindings
WETH_ABI := weth/abi/WETH.json
WETH_BINDING := weth/bindings/weth_gen.go
ABIGEN := go run github.com/ethereum/go-ethereum/cmd/abigen

help:
	@printf "Available targets:\n"
	@printf "  make test              Run unit tests\n"
	@printf "  make check             Run unit tests and whitespace checks\n"
	@printf "  make anvil-base        Start an Anvil fork of Base mainnet (requires BASE_RPC_URL)\n"
	@printf "  make test-integration  Run integration tests (requires BASE_RPC_URL)\n"
	@printf "  make update-aave-manifests Regenerate reviewed Base Aave manifests\n"
	@printf "  make update-contract-artifacts Refresh contracts artifacts and bindings\n"
	@printf "  make generate-contract-bindings Generate Go bindings from checked-in ABIs\n"
	@printf "  make verify-contract-artifacts Check contracts artifacts and bindings\n"
	@printf "  make generate-weth-binding Generate the WETH Go binding from its checked-in ABI\n"

test:
	go test ./...

check: test verify-contract-artifacts verify-weth-binding
	git diff --check

require-base-rpc:
	@if [ -z "$(BASE_RPC_URL)" ]; then \
		echo "BASE_RPC_URL is required"; \
		echo "Example: BASE_RPC_URL=https://mainnet.base.org make anvil-base"; \
		echo "Example: BASE_RPC_URL=http://127.0.0.1:8545 make test-integration"; \
		exit 1; \
	fi

require-contracts-dir:
	@if [ -z "$(CONTRACTS_DIR)" ]; then \
		echo "DEFI_SIMPLIFY_CONTRACTS_DIR is required"; \
		echo "Example: DEFI_SIMPLIFY_CONTRACTS_DIR=/path/to/defi-simplify-contracts make update-contract-artifacts"; \
		exit 1; \
	fi
	@if [ ! -d "$(CONTRACTS_DIR)" ]; then \
		echo "Contracts directory does not exist: $(CONTRACTS_DIR)"; \
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

update-contract-artifacts: require-contracts-dir
	cp "$(CONTRACTS_DIR)/abi/DefiSimplify7702Account.json" "$(CONTRACT_ARTIFACT_DIR)/abi/"
	cp "$(CONTRACTS_DIR)/abi/FlowAssertions.json" "$(CONTRACT_ARTIFACT_DIR)/abi/"
	cp "$(CONTRACTS_DIR)/abi/StaticCallUint256Assertions.json" "$(CONTRACT_ARTIFACT_DIR)/abi/"
	cp "$(CONTRACTS_DIR)/abi/DynamicExecution.golden.json" "$(CONTRACT_ARTIFACT_DIR)/abi/"
	cp "$(CONTRACTS_DIR)/abi/DynamicCalldataPatching.golden.json" "$(CONTRACT_ARTIFACT_DIR)/abi/"
	cp "$(CONTRACTS_DIR)/abi/BaseAaveV3DynamicStrategy.golden.json" "$(CONTRACT_ARTIFACT_DIR)/abi/"
	cp "$(CONTRACTS_DIR)/abi/CallbackExecution.golden.json" "$(CONTRACT_ARTIFACT_DIR)/abi/"
	cp "$(CONTRACTS_DIR)/abi/BaseAaveV3FlashLifecycle.golden.json" "$(CONTRACT_ARTIFACT_DIR)/abi/"
	cp "$(CONTRACTS_DIR)/deployments/base-v1.json" "$(CONTRACT_ARTIFACT_DIR)/deployments/"
	$(MAKE) generate-contract-bindings

generate-contract-bindings:
	$(ABIGEN) --abi $(CONTRACT_ARTIFACT_DIR)/abi/DefiSimplify7702Account.json --pkg bindings --type DefiSimplify7702Account --out $(CONTRACT_BINDING_DIR)/defi_simplify_7702_account_gen.go
	$(ABIGEN) --abi $(CONTRACT_ARTIFACT_DIR)/abi/FlowAssertions.json --pkg bindings --type FlowAssertions --out $(CONTRACT_BINDING_DIR)/flow_assertions_gen.go
	$(ABIGEN) --abi $(CONTRACT_ARTIFACT_DIR)/abi/StaticCallUint256Assertions.json --pkg bindings --type StaticCallUint256Assertions --out $(CONTRACT_BINDING_DIR)/static_call_uint256_assertions_gen.go

verify-contract-artifacts: generate-contract-bindings
	git diff --exit-code -- $(CONTRACT_BINDING_DIR)

generate-weth-binding:
	mkdir -p $(dir $(WETH_BINDING))
	$(ABIGEN) --abi $(WETH_ABI) --pkg bindings --type WETH --out $(WETH_BINDING)

verify-weth-binding: generate-weth-binding
	git diff --exit-code -- $(WETH_BINDING)
