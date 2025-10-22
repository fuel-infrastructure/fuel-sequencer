#!/usr/bin/make -f

# Sequencer's Docker image and container names.
DOCKER := $(shell which docker)
DOCKER_IMAGE_NAME := "fuel-infrastructure/fuel-sequencer"
DOCKER_IMAGE_TAG := $(shell git rev-parse --short HEAD)
DOCKER_CONTAINER_NAME := "fuel-sequencer-container"

# Fuel Rollup local repository.
ROLLUP_DIR = $(CURDIR)/e2e/fuel-rollup

# Proxy configuration
PROXY_API_PORT := 8443
PROXY_RPC_PORT := 8658
PROXY_CONTAINER_NAME := fuel-sequencer-proxy

# Name of the Ethereum contract deployment image.
ETH_DEPLOYMENT_DOCKER_IMAGE_NAME := "fuel-rollup/ethereum-deployment:latest"

# Ethereum containers' names when they are run from the E2E tests.
ETH_NODE_DOCKER_CONTAINER_NAME := "ethereum-node"
ETH_DEPLOYMENT_DOCKER_CONTAINER_NAME := "ethereum-deployment"
OTTERSCAN_DOCKER_CONTAINER_NAME := "otterscan"
# Ethereum containers' names when they are run from the docker-compose.
ETH_NODE_DOCKER_CONTAINER_NAME_COMPOSE := "eth_node"
ETH_DEPLOYMENT_DOCKER_CONTAINER_NAME_COMPOSE := "deploy"

BRANCH := $(shell git rev-parse --abbrev-ref HEAD)
COMMIT := $(shell git log -1 --format='%H')

MOCKS_DIR = $(CURDIR)/tests/mocks

# don't override user values
ifeq (,$(VERSION))
  VERSION := $(shell echo $(shell git describe --tags 2>/dev/null) | sed 's/^v//')
  # if VERSION is empty, then populate it with branch's name and raw commit hash
  ifeq (,$(VERSION))
    VERSION := $(BRANCH)-$(COMMIT)
  endif
endif

LEDGER_ENABLED ?= true

GO_SYSTEM_VERSION = $(shell go version | cut -c 14- | cut -d' ' -f1)
# Extract Go version from go.mod file
GO_MOD_VERSION = $(shell grep '^go ' go.mod | cut -d' ' -f2)
REQUIRE_GO_VERSION = $(GO_MOD_VERSION)
# Use specific Go version if available, otherwise use default go
GO_CMD = $(shell which go$(REQUIRE_GO_VERSION) 2>/dev/null || echo "go")

SDK_PACK := $(shell $(GO_CMD) list -m github.com/cosmos/cosmos-sdk | sed  's/ /\@/g')
COMETBFT_VERSION := $(shell $(GO_CMD) list -m github.com/cometbft/cometbft | sed 's:.* ::') # grab everything after the space in e.g. "github.com/cometbft/cometbft v0.37.1"
BUILDFOLDER := build
BUILDDIR ?= $(CURDIR)/$(BUILDFOLDER)

# Alpine version for Docker images (can be updated to latest stable)
ALPINE_VERSION = 3.22

export GO111MODULE = on

# process build tags

build_tags = netgo
ifeq ($(LEDGER_ENABLED),true)
  ifeq ($(OS),Windows_NT)
    GCCEXE = $(shell where gcc.exe 2> NUL)
    ifeq ($(GCCEXE),)
      $(error gcc.exe not installed for ledger support, please install or set LEDGER_ENABLED=false)
    else
      build_tags += ledger
    endif
  else
    UNAME_S = $(shell uname -s)
    ifeq ($(UNAME_S),OpenBSD)
      $(warning OpenBSD detected, disabling ledger support (https://github.com/cosmos/cosmos-sdk/issues/1988))
    else
      GCC = $(shell command -v gcc 2> /dev/null)
      ifeq ($(GCC),)
        $(error gcc not installed for ledger support, please install or set LEDGER_ENABLED=false)
      else
        build_tags += ledger
      endif
    endif
  endif
endif

ifeq (cleveldb,$(findstring cleveldb,$(FS_BUILD_OPTIONS)))
  build_tags += gcc cleveldb
endif
build_tags += $(BUILD_TAGS)
build_tags := $(strip $(build_tags))

whitespace :=
whitespace := $(whitespace) $(whitespace)
comma := ,
build_tags_comma_sep := $(subst $(whitespace),$(comma),$(build_tags))

# process linker flags

ldflags = -X github.com/cosmos/cosmos-sdk/version.Name=fuelsequencer \
		  -X github.com/cosmos/cosmos-sdk/version.AppName=fuelsequencerd \
		  -X github.com/cosmos/cosmos-sdk/version.Version=$(VERSION) \
		  -X github.com/cosmos/cosmos-sdk/version.Commit=$(COMMIT) \
		  -X "github.com/cosmos/cosmos-sdk/version.BuildTags=$(build_tags_comma_sep)" \
		  -X github.com/cometbft/cometbft/version.TMCoreSemVer=$(COMETBFT_VERSION)

ifeq (cleveldb,$(findstring cleveldb,$(FS_BUILD_OPTIONS)))
  ldflags += -X github.com/cosmos/cosmos-sdk/types.DBBackend=cleveldb
endif
ifeq (,$(findstring nostrip,$(FS_BUILD_OPTIONS)))
  ldflags += -w -s
endif
ldflags += $(LDFLAGS)
ldflags := $(strip $(ldflags))

BUILD_FLAGS := -tags "$(build_tags)" -ldflags '$(ldflags)'
# check for nostrip option
ifeq (,$(findstring nostrip,$(FS_BUILD_OPTIONS)))
  BUILD_FLAGS += -trimpath
endif

###############################################################################
###                              Build & Clean                              ###
###############################################################################

check-go-version:
	@if echo "$(GO_CMD)" | grep -q "go$(REQUIRE_GO_VERSION)$$"; then \
		echo "✅ Found go$(REQUIRE_GO_VERSION) command, using it for build"; \
	elif [ "$(GO_SYSTEM_VERSION)" = "$(REQUIRE_GO_VERSION)" ]; then \
		echo "✅ Go version requirements are met for $(VERSION) of FuelSequencer."; \
	else \
		echo "❌ Go version $(REQUIRE_GO_VERSION) is required for $(VERSION) of FuelSequencer."; \
		echo "   Either install go$(REQUIRE_GO_VERSION) or ensure your current go version is $(REQUIRE_GO_VERSION)"; \
		echo "   Refer https://go.dev/doc/manage-install to install go$(REQUIRE_GO_VERSION)"; \
		exit 1; \
	fi

all: install lint test-unit vulncheck

BUILD_TARGETS := build install

build: BUILD_ARGS=-o $(BUILDDIR)/

$(BUILD_TARGETS): check-go-version go.sum $(BUILDDIR)/
	@$(GO_CMD) $@ -mod=readonly $(BUILD_FLAGS) $(BUILD_ARGS) ./...

$(BUILDDIR)/:
	mkdir -p $(BUILDDIR)/

vulncheck: $(BUILDDIR)/
	GOBIN=$(BUILDDIR) $(GO_CMD) install golang.org/x/vuln/cmd/govulncheck@latest
	$(BUILDDIR)/govulncheck ./...

build-linux: go.sum
	LEDGER_ENABLED=false GOOS=linux GOARCH=amd64 $(MAKE) build

go-mod-cache: go.sum
	@echo "⬇️ Download go modules to local cache"
	@$(GO_CMD) mod download

go.sum: go.mod
	@echo "🤔 Ensure dependencies have not been modified"
	@$(GO_CMD) mod verify

clean: clean-e2e
	@echo "🧹 Cleaning..."
	@rm -rf $(BUILDDIR)/*
	@echo "✅ Finished cleaning!"

build-fuelsequencerd:
	@$(eval UNAME_S := $(shell uname -s))
    ifeq ($(UNAME_S),Darwin)
		@echo "⚠️ Only building darwin binaries. Linux device required to build linux binaries."
		@echo "🔧 Building fuelsequencerd-$(VERSION)-darwin-amd64..."
		@GOOS=darwin GOARCH=amd64 $(GO_CMD) build -mod=readonly $(BUILD_FLAGS) -o $(BUILDDIR)/fuelsequencerd-$(VERSION)-darwin-amd64 ./cmd/fuelsequencerd/main.go
		@echo "🔧 Building fuelsequencerd-$(VERSION)-darwin-arm64..."
		@GOOS=darwin GOARCH=arm64 $(GO_CMD) build -mod=readonly $(BUILD_FLAGS) -o $(BUILDDIR)/fuelsequencerd-$(VERSION)-darwin-arm64 ./cmd/fuelsequencerd/main.go
    else
		@echo "⚠️ Only building linux binaries. Darwin device required to build darwin binaries."
		@echo "🔧 Building fuelsequencerd-$(VERSION)-linux-amd64..."
		@GOOS=linux GOARCH=amd64 $(GO_CMD) build -mod=readonly $(BUILD_FLAGS) -o $(BUILDDIR)/fuelsequencerd-$(VERSION)-linux-amd64 ./cmd/fuelsequencerd/main.go
		@echo "🔧 Building fuelsequencerd-$(VERSION)-linux-arm64..."
		@GOOS=linux GOARCH=arm64 $(GO_CMD) build -mod=readonly $(BUILD_FLAGS) -o $(BUILDDIR)/fuelsequencerd-$(VERSION)-linux-arm64 ./cmd/fuelsequencerd/main.go
    endif

build-all: clean build-fuelsequencerd
	@echo "✅ Finished building all!"

do-checksum:
	@echo "🤖 Generating checksum..."
	@cd $(BUILDDIR)/ && find ./fuelsequencerd* -type f -exec sha256sum {} \; > sha256sum.txt
	@echo "✅ Finished generating checksum!"

build-with-checksum: build-all do-checksum

run-client-binary:
	@$(eval ARCH ?= linux-amd64)
	@if [ -z "$(BLOCK_NUMBER)" ]; then echo "BLOCK_NUMBER is not set. Use make run-client BLOCK_NUMBER=<number>"; exit 1; fi
	@echo "Running client $(VERSION) for $(ARCH) with block number $(BLOCK_NUMBER)..."
	@$(BUILDDIR)/client-$(VERSION)-$(ARCH) -blocknumber $(BLOCK_NUMBER)

run-sidecar-binary:
	@$(eval SIDECAR_HOST ?= "0.0.0.0")
	@$(eval SIDECAR_PORT ?= "8080")
	@$(eval SIDECAR_PATH_TO_CERT_FILE ?= "")
	@$(eval SIDECAR_PATH_TO_KEY_FILE ?= "")
	@$(eval SEQUENCER_GRPC_URL ?= "127.0.0.1:9090")
	@$(eval SEQUENCER_PATH_TO_CERT_FILE ?= "")
	@$(eval UNSAFE_SEQUENCER_BRIDGE_DENOM ?= "utest")
	@$(eval ETH_WS_URL ?= "ws://localhost:8545")
	@$(eval ETH_RPC_URL ?= "http://localhost:8545")
	@$(eval ETH_CONTRACT_ADDRESS ?= "0x0165878A594ca255338adfa4d48449f69242Eb8F")
	@$(eval ETH_MAX_BLOCK_RANGE ?= "100")
	@$(eval ETH_MIN_LOGS_QUERY_INTERVAL ?= "10s")
	@$(eval ETH_RPC_QUERY_TIMEOUT ?= "20s")
	@$(eval DEVELOPMENT ?= "false")
	@$(eval ARCH ?= linux-amd64)
	@echo "Running sidecar $(VERSION) for $(ARCH)..."
	@$(BUILDDIR)/sidecar-$(VERSION)-$(ARCH) \
		--host "$(SIDECAR_HOST)" \
		--port "$(SIDECAR_PORT)" \
		--sidecar_path_to_cert_file "$(SIDECAR_PATH_TO_CERT_FILE)" \
		--sidecar_path_to_key_file "$(SIDECAR_PATH_TO_KEY_FILE)" \
		--sequencer_grpc_url "$(SEQUENCER_GRPC_URL)" \
		--sequencer_path_to_cert_file "$(SEQUENCER_PATH_TO_CERT_FILE)" \
		--unsafe_sequencer_bridge_denom "$(UNSAFE_SEQUENCER_BRIDGE_DENOM)" \
		--eth_ws_url "$(ETH_WS_URL)" \
		--eth_rpc_url "$(ETH_RPC_URL)" \
		--eth_contract_address "$(ETH_CONTRACT_ADDRESS)" \
		--eth_max_block_range "$(ETH_MAX_BLOCK_RANGE)" \
		--eth_min_logs_query_interval "$(ETH_MIN_LOGS_QUERY_INTERVAL)" \
		--eth_rpc_query_timeout "$(ETH_RPC_QUERY_TIMEOUT)" \
		--development "$(DEVELOPMENT)"

###############################################################################
###                                 Protobuf                                ###
###############################################################################

containerProtoVer=v0.14.0
containerProtoFmt=cosmos-sdk-proto-fmt-$(containerProtoVer)

protoVer=0.14.0
protoImageName=ghcr.io/cosmos/proto-builder:$(protoVer)
protoImage=$(DOCKER) run --rm -v $(CURDIR):/workspace --workdir /workspace $(protoImageName)

cosmos_sdk_dir=$(shell $(GO_CMD) list -f '{{ .Dir }}' -m github.com/cosmos/cosmos-sdk)
protoSwaggerImage=$(DOCKER) run --rm -v $(CURDIR):/workspace -v $(cosmos_sdk_dir):/cosmos-sdk --workdir /workspace $(protoImageName)

proto-go-gen:
    # This runs ./scripts/protocgen-pulsar.sh as well, under the hood.
	@echo "🤖 Generating Go code from protobuf..."
	@$(protoImage) sh ./scripts/protocgen.sh;
	@echo "✅ Finished Go code generation!"

proto-format:
	@echo "🤖 Formatting Protobuf files..."
	@$(DOCKER) run --rm --name $(containerProtoFmt) -v $(CURDIR):/workspace --workdir /workspace tendermintdev/docker-build-proto \
		find ./proto -name "*.proto" -exec clang-format -i {} \;
	@echo "✅ Finished formatting Protobuf files!"

proto-swagger-gen:
	@echo "🤖 Generating API docs..."
	@$(protoSwaggerImage) sh ./scripts/protoc-swagger-gen.sh

proto-routine: proto-format proto-go-gen proto-swagger-gen

###############################################################################
###                                   Run                                   ###
###############################################################################

run-sequencer: proto-go-gen serve

run-sidecar:
	@$(eval SIDECAR_HOST ?= "0.0.0.0")
	@$(eval SIDECAR_PORT ?= "8080")
	@$(eval SIDECAR_PATH_TO_CERT_FILE ?= "")
	@$(eval SIDECAR_PATH_TO_KEY_FILE ?= "")
	@$(eval SEQUENCER_GRPC_URL ?= "127.0.0.1:9090")
	@$(eval SEQUENCER_RPC_URL ?= "http://127.0.0.1:26657")  # for the wait below
	@$(eval SEQUENCER_PATH_TO_CERT_FILE ?= "")
	@$(eval UNSAFE_SEQUENCER_BRIDGE_DENOM ?= "utest")
	@$(eval ETH_WS_URL ?= "ws://localhost:8545")
	@$(eval ETH_RPC_URL ?= "http://localhost:8545")  # for the wait below and Sidecar RPC calls
	@$(eval ETH_CONTRACT_ADDRESS ?= "0x0165878A594ca255338adfa4d48449f69242Eb8F")
	@$(eval ETH_MAX_BLOCK_RANGE ?= "1")
	@$(eval ETH_MIN_LOGS_QUERY_INTERVAL ?= "1s")
	@$(eval ETH_RPC_QUERY_TIMEOUT ?= "20s")
	@$(eval DEVELOPMENT ?= "true")
	@$(eval PROMETHEUS_ENABLED ?= "true")
	@echo "Waiting for Ethereum node $(ETH_RPC_URL) to start..."
	@while ! curl -s -X POST -H "Content-Type: application/json" --data '{"jsonrpc":"2.0","method":"web3_clientVersion","params":[],"id":1}' --max-time 1 $(ETH_RPC_URL) | grep -q "result"; do \
	    sleep 1; \
	done
	@echo "Waiting for Ethereum deployment container '$(ETH_DEPLOYMENT_DOCKER_CONTAINER_NAME_COMPOSE)' to stop..."
	@while [ -n "$$($(DOCKER) ps -q -f name=$(ETH_DEPLOYMENT_DOCKER_CONTAINER_NAME_COMPOSE))" ]; do \
		sleep 1; \
	done
	@echo "Waiting for Sequencer node $(SEQUENCER_RPC_URL) to start..."
	@while ! curl -s -X GET --max-time 1 "$(SEQUENCER_RPC_URL)" | grep -q "result"; do \
		sleep 1; \
	done
	@echo "Waiting for Sequencer gRPC $(SEQUENCER_GRPC_URL) to be accessible..."
	@sleep 3  # buffer for Sequencer gRPC server to start properly
	@fuelsequencerd start-sidecar \
		--host="$(SIDECAR_HOST)" \
		--port="$(SIDECAR_PORT)" \
		--sequencer_grpc_url="$(SEQUENCER_GRPC_URL)" \
		--sequencer_path_to_cert_file="$(SEQUENCER_PATH_TO_CERT_FILE)" \
		--unsafe_sequencer_bridge_denom="$(UNSAFE_SEQUENCER_BRIDGE_DENOM)" \
		--sidecar_path_to_cert_file="$(SIDECAR_PATH_TO_CERT_FILE)" \
		--sidecar_path_to_key_file="$(SIDECAR_PATH_TO_KEY_FILE)" \
		--eth_ws_url="$(ETH_WS_URL)" \
		--eth_rpc_url="$(ETH_RPC_URL)" \
		--eth_contract_address="$(ETH_CONTRACT_ADDRESS)" \
		--eth_max_block_range="$(ETH_MAX_BLOCK_RANGE)" \
		--eth_min_logs_query_interval="$(ETH_MIN_LOGS_QUERY_INTERVAL)" \
		--eth_rpc_query_timeout="$(ETH_RPC_QUERY_TIMEOUT)" \
		--development="$(DEVELOPMENT)" \
		--prometheus_enabled="$(PROMETHEUS_ENABLED)"

init:
	ignite chain init --skip-proto --build.tags ledger

serve:
	ignite chain serve -v --reset-once --skip-proto --build.tags ledger

serve-force-reset:
	ignite chain serve -v --force-reset --skip-proto --build.tags ledger

keys:
	@echo "🤖 Generating keys..."

	@$(eval MNEMONIC := "dinner crash nurse casino baby fold race cheese elite column sausage sleep close royal rain over mechanic minimum outdoor conduct cash wagon frog evidence")
	@- fuelsequencerd keys delete alice -y
	yes $(MNEMONIC) | fuelsequencerd keys add alice --recover

	@$(eval MNEMONIC := "gaze drama excess raven follow antenna swallow beef upper myself question pitch course ill adult century crisp ice rough match praise sing unveil vintage")
	@- fuelsequencerd keys delete bob -y
	@yes $(MNEMONIC) | fuelsequencerd keys add bob --recover

	@$(eval MNEMONIC := "bar describe panda mosquito quiz room daring round nurse disagree swallow frown hat repeat recall flight skin sketch volume dutch range grunt assist nerve")
	@- fuelsequencerd keys delete carol -y
	@yes $(MNEMONIC) | fuelsequencerd keys add carol --recover

	@$(eval MNEMONIC := "bonus clinic owner choose grief soda ride divorce album oval tone mixed mechanic coin defense wonder tumble vault sorry great hover neither security amazing")
	@- fuelsequencerd keys delete dexter -y
	@yes $(MNEMONIC) | fuelsequencerd keys add dexter --recover

	@echo "✅ Finished generating keys!"


###############################################################################
###                                   CI                                    ###
###############################################################################

ci: proto-routine mocks format lint test-unit

lint:
	@echo "🔎 Running linter..."
	@golangci-lint run --timeout=10m --fix
	@echo "🔎 Running linter for e2e..."
	@cd e2e && golangci-lint run --timeout=10m --fix && cd ..
	@echo "✅ Finished running linter!"

# Extract and convert excluded paths from .golangci.yml to find-compatible patterns
# 1. Extract lines with 'path:' that end in '$'
# 2. Remove 'path:' and whitespace
# 3. Remove regex end marker '$'
# 4. Convert '.ext.go$' pattern to '-not -name "*.ext.go"'
EXCLUDED_PATTERNS := $(shell awk '/path:.*\.go\$$/ { \
		gsub(/.*path: /, ""); \
		gsub(/\$$/, ""); \
		gsub(/\\\./, "."); \
		printf "-not -name \"*%s\" ", $$0 \
	}' .golangci.yml)

format:
	@echo "🔎 Running formatter..."
	@find . -type f -name "*.go" $(EXCLUDED_PATTERNS) \
		| xargs goimports -w -local $(shell $(GO_CMD) list -m)
	@echo "✅ Finished running formatter!"

###############################################################################
###                                  Tests                                  ###
###############################################################################

test-all: test-unit test-e2e

test-unit:
	@echo "🤖 Running unit tests..."
	@$(GO_CMD) test -mod=readonly ./x/$(module)/... ./sidecar/... ./app/...

test-coverage:
	@echo "🤖 Generating test coverage..."
	@mkdir -p $(BUILDDIR)/coverage
	@$(GO_CMD) test -mod=readonly -coverprofile=$(BUILDDIR)/coverage/coverage.out ./x/... ./app/... ./sidecar/...
	@$(GO_CMD) tool cover -html=$(BUILDDIR)/coverage/coverage.out -o $(BUILDDIR)/coverage/coverage.html
	@$(GO_CMD) tool cover -func=$(BUILDDIR)/coverage/coverage.out
	@echo "✅ Coverage report generated at $(BUILDDIR)/coverage/coverage.html"

open-coverage:
	@echo "🌐 Opening coverage report..."
	@open $(BUILDDIR)/coverage/coverage.html

test-e2e: \
	check-docker-image-exists \
	check-eth-deployment-docker-image-exists \
	test-e2e-basic \
	test-e2e-withdrawals \
	test-e2e-events \
	test-e2e-authorize \
	test-e2e-deposits \
	test-e2e-special-messages

test-cover:
	@$(GO_CMD) test -mod=readonly -race -coverprofile=coverage.out -covermode=atomic ./x/$(module)/... ./sidecar/... ./app/...

test-e2e-vulnerability-mitigations:
	@cd e2e/tests && $(GO_CMD) test -mod=readonly -race -v ./upgrade/vulnerability-mitigations/... --test.timeout 0

mocks: $(MOCKS_DIR)
	@$(GO_CMD) install github.com/golang/mock/mockgen@v1.6.0
	sh ./scripts/mockgen.sh
	rm -r "$(MOCKS_DIR)"
.PHONY: mocks

$(MOCKS_DIR):
	mkdir -p $(MOCKS_DIR)

###############################################################################
###                                 Metrics                                 ###
###############################################################################

#? metrics: Generate metrics
metrics:
	$(GO_CMD) generate -run="scripts/metricsgen" ./...
.PHONY: metrics

###############################################################################
###                                Docker                                   ###
###############################################################################

check-docker-image-exists:
ifeq (,$(shell $(DOCKER) images -q ${DOCKER_IMAGE_NAME}:latest 2> /dev/null))
	@echo "❌ Docker image ${DOCKER_IMAGE_NAME}:latest not found";
	@exit 1;
else
	@echo "✅ Found docker image ${DOCKER_IMAGE_NAME}:latest"
endif

build-docker-image:
	@echo "🤖 Building Docker image..."
	@$(DOCKER) build \
		-t ${DOCKER_IMAGE_NAME}:${DOCKER_IMAGE_TAG} \
		--build-arg GO_VERSION=${REQUIRE_GO_VERSION} \
		--build-arg ALPINE_VERSION=${ALPINE_VERSION} \
		.
	@$(DOCKER) tag ${DOCKER_IMAGE_NAME}:${DOCKER_IMAGE_TAG} ${DOCKER_IMAGE_NAME}:$(shell echo ${BRANCH} | sed 's|/|_|g')
	@echo Successfully tagged ${DOCKER_IMAGE_NAME}:$(shell echo ${BRANCH} | sed 's|/|_|g')
	@$(DOCKER) tag ${DOCKER_IMAGE_NAME}:${DOCKER_IMAGE_TAG} ${DOCKER_IMAGE_NAME}:latest
	@echo Successfully tagged ${DOCKER_IMAGE_NAME}:latest
	@echo "✅ Finished building Docker image!"

run-docker-container: check-docker-image-exists
	@if [ -z "$(ETH_RPC_URL)" ]; then \
		echo "ETH_RPC_URL is not set"; \
		exit 1; \
	fi
	@if [ -z "$(ETH_WS_URL)" ]; then \
		echo "ETH_WS_URL is not set"; \
		exit 1; \
	fi
	@$(eval DATA_FOLDER ?= "/data/fuelsequencer")
	@$(eval COMMAND ?= "node_and_sidecar")
	@echo "🤖 Running Docker container..."
	@$(DOCKER) run -d \
    		-v $(shell pwd)${DATA_FOLDER}:/home/fuelsequencer/.fuelsequencer \
    		--name $(DOCKER_CONTAINER_NAME) \
    		-p 26656:26656 -p 26657:26657 -p 1317:1317 -p 8080:8080 -p 8081:8081 \
    		-e "ETH_RPC_URL=$(ETH_RPC_URL)" \
    		-e "ETH_WS_URL=$(ETH_WS_URL)" \
    		${DOCKER_IMAGE_NAME}:latest \
    		${COMMAND}

start-docker-container:
	@echo "🤖 Starting Docker container..."
	@$(DOCKER) start $(DOCKER_CONTAINER_NAME)
	@echo "✅ Started Docker container!"

stop-docker-container:
	@echo "🤖 Stopping Docker container..."
	@$(DOCKER) stop $(DOCKER_CONTAINER_NAME)
	@echo "✅ Stopped Docker container!"

remove-docker-container:
	@echo "🤖 Removing Docker container..."
	@$(DOCKER) rm -v $(DOCKER_CONTAINER_NAME)
	@echo "✅ Removed Docker container!"

follow-docker-logs:
	@$(DOCKER) logs -f $(DOCKER_CONTAINER_NAME)

###############################################################################
###                                   E2E                                   ###
###############################################################################

build-all-docker-images: \
	build-docker-image \
	build-eth-deployment-docker-image

check-eth-deployment-docker-image-exists:
ifeq (,$(shell $(DOCKER) images -q ${ETH_DEPLOYMENT_DOCKER_IMAGE_NAME} 2> /dev/null))
	@echo "❌ Docker image ${ETH_DEPLOYMENT_DOCKER_IMAGE_NAME} not found";
	@exit 1;
else
	@echo "✅ Found docker image ${ETH_DEPLOYMENT_DOCKER_IMAGE_NAME}"
endif

# Builds contract deployment container for automated E2E tests
# Note: this assumes evm_setIntervalMining is set to 3.
build-eth-deployment-docker-image: $(ROLLUP_DIR)/.npmrc
	@echo "🤖 Updating git submodules (fuel-rollup)..."
	@git submodule update --init --remote $(ROLLUP_DIR)
	@echo "🤖 Building dependencies..."
	@(cd $(ROLLUP_DIR) && pnpm install && pnpm build)
	@echo "🤖 Building Docker image..."
	@$(DOCKER) build \
		-t $(ETH_DEPLOYMENT_DOCKER_IMAGE_NAME) \
		-f $(ROLLUP_DIR)/docker/docker.eth_node.Dockerfile \
		--build-arg NPM_TOKEN=$$NPM_TOKEN \
		$(ROLLUP_DIR)
	@echo "🤖 Cleaning up git submodules (fuel-rollup)..."
	@git submodule update --remote $(ROLLUP_DIR)
	@echo "✅ Finished!"

# Runs node and contract deployment containers
run-eth-e2e-containers: $(ROLLUP_DIR)/.npmrc
	@echo "🤖 Running Docker containers..."
	@$(DOCKER) compose -f $(ROLLUP_DIR)/docker/docker-compose.yml up -d --build \
		"$(ETH_NODE_DOCKER_CONTAINER_NAME_COMPOSE)" \
		"$(ETH_DEPLOYMENT_DOCKER_CONTAINER_NAME_COMPOSE)"

# Removes node and contract deployment containers
remove-eth-e2e-containers:
	@echo "🤖 Removing Docker containers..."
	@$(DOCKER) compose -f $(ROLLUP_DIR)/docker/docker-compose.yml down
	@echo "✅ Removed Docker containers!"

test-e2e-basic:
	@cd e2e/tests && $(GO_CMD) test -mod=readonly -race -v ./basic/... --test.timeout 0

test-e2e-events:
	@cd e2e/tests && $(GO_CMD) test -mod=readonly -race -v ./events/... --test.timeout 0

test-e2e-withdrawals:
	@cd e2e/tests && $(GO_CMD) test -mod=readonly -race -v ./withdrawals/... --test.timeout 0

test-e2e-authorize:
	@cd e2e/tests && $(GO_CMD) test -mod=readonly -race -v ./authorize/... --test.timeout 0

test-e2e-deposits:
	@cd e2e/tests && $(GO_CMD) test -mod=readonly -race -v ./deposits/... --test.timeout 0

test-e2e-special-messages:
	@cd e2e/tests && $(GO_CMD) test -mod=readonly -race -v ./special-messages/... --test.timeout 0

# Run e2e tests with HTTPS proxy for fuel-explorer integration
test-e2e-with-proxy: check-docker-image-exists
	@echo "🔐 Running E2E tests with integrated HTTPS proxy"
	@echo "📊 Endpoints available during tests:"
	@echo "  • API (HTTPS):  https://localhost:$(PROXY_API_PORT)"
	@echo "  • RPC (HTTPS):  https://localhost:$(PROXY_RPC_PORT)"
	@echo "🔒 Note: Self-signed certificates will be used"
	@cd e2e/tests && $(GO_CMD) test -mod=readonly -race -v ./proxy/... --test.timeout 0

# Run e2e tests with HTTPS proxy for fuel-explorer integration
test-e2e-with-proxy: check-docker-image-exists
	@echo "🔐 Running E2E tests with integrated HTTPS proxy"
	@echo "📊 Endpoints available during tests:"
	@echo "  • API (HTTPS):  https://localhost:$(PROXY_API_PORT)"
	@echo "  • RPC (HTTPS):  https://localhost:$(PROXY_RPC_PORT)"
	@echo "🔒 Note: Self-signed certificates will be used"
	@cd e2e/tests && go test -mod=readonly -race -v ./proxy/... --test.timeout 0

clean-e2e:
	@echo "🧹 Stopping Docker containers..."
	@$(DOCKER) ps -aq --filter "name=fuelsequencer0" | xargs -r $(DOCKER) stop
	@$(DOCKER) ps -aq --filter "name=fuelsequencer1" | xargs -r $(DOCKER) stop
	@$(DOCKER) ps -aq --filter "name=fuelsequencer2" | xargs -r $(DOCKER) stop
	@$(DOCKER) ps -aq --filter "name=$(ETH_NODE_DOCKER_CONTAINER_NAME)" | xargs -r $(DOCKER) stop
	@$(DOCKER) ps -aq --filter "name=$(ETH_DEPLOYMENT_DOCKER_CONTAINER_NAME)" | xargs -r $(DOCKER) stop
	@$(DOCKER) ps -aq --filter "name=$(PROXY_CONTAINER_NAME)" | xargs -r $(DOCKER) stop
	@$(DOCKER) compose -f $(ROLLUP_DIR)/docker/docker-compose.yml down

	@echo "🧹 Removing Docker containers..."
	@$(DOCKER) ps -aq --filter "name=fuelsequencer0" | xargs -r $(DOCKER) rm
	@$(DOCKER) ps -aq --filter "name=fuelsequencer1" | xargs -r $(DOCKER) rm
	@$(DOCKER) ps -aq --filter "name=fuelsequencer2" | xargs -r $(DOCKER) rm
	@$(DOCKER) ps -aq --filter "name=$(ETH_NODE_DOCKER_CONTAINER_NAME)" | xargs -r $(DOCKER) rm
	@$(DOCKER) ps -aq --filter "name=$(ETH_DEPLOYMENT_DOCKER_CONTAINER_NAME)" | xargs -r $(DOCKER) rm
	@$(DOCKER) ps -aq --filter "name=$(PROXY_CONTAINER_NAME)" | xargs -r $(DOCKER) rm
	@$(DOCKER) ps -aq --filter "name=$(OTTERSCAN_DOCKER_CONTAINER_NAME)" | xargs -r $(DOCKER) rm

	@echo "🧹 Pruning Docker networks..."
	@$(DOCKER) network prune -f

	@echo "✅ Finished cleaning E2E!"
