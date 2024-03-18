#!/usr/bin/make -f

DOCKER := $(shell which docker)
DOCKER_IMAGE_NAME := "fuel-infrastructure/fuel-sequencer"
DOCKER_IMAGE_TAG := $(shell git rev-parse --short HEAD)
DOCKER_CONTAINER_NAME := "fuel-sequencer-container"

ETH_DOCKER_IMAGE_NAME := "fuel-infrastructure/contracts-docker-e2e"

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
SDK_PACK := $(shell go list -m github.com/cosmos/cosmos-sdk | sed  's/ /\@/g')
COMETBFT_VERSION := $(shell go list -m github.com/cometbft/cometbft | sed 's:.* ::') # grab everything after the space in e.g. "github.com/cometbft/cometbft v0.37.1"
BUILDFOLDER := build
BUILDDIR ?= $(CURDIR)/$(BUILDFOLDER)

GO_SYSTEM_VERSION = $(shell go version | cut -c 14- | cut -d' ' -f1 | cut -d'.' -f1-2)
REQUIRE_GO_VERSION = 1.21

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
ifneq ($(GO_SYSTEM_VERSION), $(REQUIRE_GO_VERSION))
	@echo "❌ Go version $(REQUIRE_GO_VERSION) is required for $(VERSION) of FuelSequencer."
	@exit 1
else
	@echo "✅ Go version requirements are met for $(VERSION) of FuelSequencer."
endif

all: install lint test-unit vulncheck

BUILD_TARGETS := build install

build: BUILD_ARGS=-o $(BUILDDIR)/

$(BUILD_TARGETS): check-go-version go.sum $(BUILDDIR)/
	go $@ -mod=readonly $(BUILD_FLAGS) $(BUILD_ARGS) ./...

$(BUILDDIR)/:
	mkdir -p $(BUILDDIR)/

vulncheck: $(BUILDDIR)/
	GOBIN=$(BUILDDIR) go install golang.org/x/vuln/cmd/govulncheck@latest
	$(BUILDDIR)/govulncheck ./...

build-linux: go.sum
	LEDGER_ENABLED=false GOOS=linux GOARCH=amd64 $(MAKE) build

go-mod-cache: go.sum
	@echo "⬇️ Download go modules to local cache"
	@go mod download

go.sum: go.mod
	@echo "🤔 Ensure dependencies have not been modified"
	@go mod verify

clean:
	@echo "🧹 Cleaning..."
	@rm -rf $(BUILDDIR)/*
	@echo "✅ Finished cleaning!"

build-fuelsequencerd:
	@$(eval MAIN := ./cmd/fuelsequencerd/main.go)
	@echo "🔧 Building fuelsequencerd-$(VERSION)-linux-amd64..."
	@GOOS=linux GOARCH=amd64 go build -mod=readonly $(BUILD_FLAGS) -o $(BUILDDIR)/fuelsequencerd-$(VERSION)-linux-amd64 $(MAIN)
	
	@echo "🔧 Building fuelsequencerd-$(VERSION)-linux-arm64..."
	@GOOS=linux GOARCH=arm64 go build -mod=readonly $(BUILD_FLAGS) -o $(BUILDDIR)/fuelsequencerd-$(VERSION)-linux-arm64 $(MAIN)

	@echo "🔧 Building fuelsequencerd-$(VERSION)-darwin-amd64..."
	@GOOS=darwin GOARCH=amd64 go build -mod=readonly $(BUILD_FLAGS) -o $(BUILDDIR)/fuelsequencerd-$(VERSION)-darwin-amd64 $(MAIN)

build-sidecar:
	@$(eval MAIN := ./cmd/sidecar/main.go)
	@echo "🔧 Building sidecar-$(VERSION)-linux-amd64..."
	@GOOS=linux GOARCH=amd64 go build -mod=readonly $(BUILD_FLAGS) -o $(BUILDDIR)/sidecar-$(VERSION)-linux-amd64 $(MAIN)
	@echo "🔧 Building sidecar-$(VERSION)-linux-arm64..."
	@GOOS=linux GOARCH=arm64 go build -mod=readonly $(BUILD_FLAGS) -o $(BUILDDIR)/sidecar-$(VERSION)-linux-arm64 $(MAIN)
	@echo "🔧 Building sidecar-$(VERSION)-darwin-amd64..."
	@GOOS=darwin GOARCH=amd64 go build -mod=readonly $(BUILD_FLAGS) -o $(BUILDDIR)/sidecar-$(VERSION)-darwin-amd64 $(MAIN)

build-client:
	@$(eval MAIN := ./cmd/client/main.go)
	@echo "🔧 Building client-$(VERSION)-linux-amd64..."
	@GOOS=linux GOARCH=amd64 go build -mod=readonly $(BUILD_FLAGS) -o $(BUILDDIR)/client-$(VERSION)-linux-amd64 $(MAIN)
	@echo "🔧 Building client-$(VERSION)-linux-arm64..."
	@GOOS=linux GOARCH=arm64 go build -mod=readonly $(BUILD_FLAGS) -o $(BUILDDIR)/client-$(VERSION)-linux-arm64 $(MAIN)
	@echo "🔧 Building client-$(VERSION)-darwin-amd64..."
	@GOOS=darwin GOARCH=amd64 go build -mod=readonly $(BUILD_FLAGS) -o $(BUILDDIR)/client-$(VERSION)-darwin-amd64 $(MAIN)

build-all: clean build-fuelsequencerd build-sidecar build-client
	@echo "✅ Finished building all!"

do-checksum:
	@echo "🤖 Generating checksum..."
	@cd $(BUILDDIR)/ && find ./fuelsequencerd* -type f -exec sha256sum {} \; > sha256sum.txt
	@echo "✅ Finished generating checksum!"

build-with-checksum: build-all do-checksum

run-client:
	@$(eval ARCH := linux-amd64)
	@if [ -z "$(BLOCK_NUMBER)" ]; then echo "BLOCK_NUMBER is not set. Use make run-client BLOCK_NUMBER=<number>"; exit 1; fi
	@echo "Running client $(VERSION) for $(ARCH) with block number $(BLOCK_NUMBER)..."
	@$(BUILDDIR)/client-$(VERSION)-$(ARCH) -blocknumber $(BLOCK_NUMBER)

run-sidecar:
	@$(eval ARCH := linux-amd64)
	@echo "Running sidecar $(VERSION) for $(ARCH)..."
	@$(BUILDDIR)/sidecar-$(VERSION)-$(ARCH) --host="$(HOST)" --port="$(PORT)" --eth_node_rpc="$(ETH_NODE_RPC)" --contract_address="$(CONTRACT_ADDRESS)" --eth_start_block="$(ETH_START_BLOCK)" --development="$(DEVELOPMENT)"

###############################################################################
###                                 Protobuf                                ###
###############################################################################

containerProtoVer=v0.14.0
containerProtoFmt=cosmos-sdk-proto-fmt-$(containerProtoVer)

protoVer=0.14.0
protoImageName=ghcr.io/cosmos/proto-builder:$(protoVer)
protoImage=$(DOCKER) run --rm -v $(CURDIR):/workspace --workdir /workspace $(protoImageName)

cosmos_sdk_dir=$(shell go list -f '{{ .Dir }}' -m github.com/cosmos/cosmos-sdk)
protoSwaggerImage=$(DOCKER) run --rm -v $(CURDIR):/workspace -v $(cosmos_sdk_dir):/cosmos-sdk --workdir /workspace $(protoImageName)

proto-go-gen:
    # This runs ./utils/protocgen-pulsar.sh as well, under the hood.
	@echo "🤖 Generating Go code from protobuf..."
	@$(protoImage) sh ./utils/protocgen.sh;
	@echo "✅ Finished Go code generation!"

proto-format:
	@echo "🤖 Formatting Protobuf files..."
	@if docker ps -a --format '{{.Names}}' | grep -Eq "^${containerProtoFmt}$$"; then docker start -a $(containerProtoFmt); else docker run --name $(containerProtoFmt) -v $(CURDIR):/workspace --workdir /workspace tendermintdev/docker-build-proto \
		find ./proto -name "*.proto" -exec clang-format -i {} \; ; fi
	@echo "✅ Finished formatting Protobuf files!"

# This command makes use of Ignite's new way of specifying docs.
# This can be improved later on with a non-Ignite approach.
proto-swagger-gen:
	@echo "🤖 Generating Swagger files..."
	@ignite generate openapi
	@echo "✅ Finished generating Swagger files!"

proto-routine: proto-format proto-go-gen proto-swagger-gen

###############################################################################
###                                   Run                                   ###
###############################################################################

run: proto-go-gen serve

serve:
	ignite chain serve --reset-once --skip-proto --build.tags ledger

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

ci: proto-routine lint test-unit gosec

gosec:
	@go run github.com/securego/gosec/v2/cmd/gosec -exclude-dir=deps -severity=high ./...

lint:
	@echo "🔎 Running linter..."
	@go run github.com/golangci/golangci-lint/cmd/golangci-lint run --timeout=10m
	@echo "✅ Finished running linter!"

###############################################################################
###                                  Tests                                  ###
###############################################################################

test-all: test-unit test-e2e

test-unit:
	@go test -mod=readonly ./x/$(module)/... ./sidecar/... ./app/...
	@# Run some unit tests that are in the e2e folder:
	@cd e2e && go test -mod=readonly -race -v ./testsuite/... --test.timeout 0

test-e2e: test-e2e-basic

test-cover:
	@go test -mod=readonly -race -coverprofile=coverage.out -covermode=atomic ./x/$(module)/... ./sidecar/... ./app/...

mocks: $(MOCKS_DIR)
	@go install github.com/golang/mock/mockgen@v1.6.0
	sh ./utils/mockgen.sh
.PHONY: mocks

$(MOCKS_DIR):
	mkdir -p $(MOCKS_DIR)

###############################################################################
###                                Docker                                   ###
###############################################################################

check-docker-image-exists:
ifeq (,$(shell docker images -q ${DOCKER_IMAGE_NAME}:latest 2> /dev/null))
	@echo "❌ Docker image ${DOCKER_IMAGE_NAME}:latest not found";
	@exit 1;
else
	@echo "✅ Found docker image ${DOCKER_IMAGE_NAME}:latest"
endif

build-docker-image:
	@echo "🤖 Building Docker image..."
	@docker build \
		-t ${DOCKER_IMAGE_NAME}:${DOCKER_IMAGE_TAG} \
		--build-arg GO_VERSION=${REQUIRE_GO_VERSION} \
		.
	@docker tag ${DOCKER_IMAGE_NAME}:${DOCKER_IMAGE_TAG} ${DOCKER_IMAGE_NAME}:$(shell echo ${BRANCH} | sed 's|/|_|g')
	@echo Successfully tagged ${DOCKER_IMAGE_NAME}:$(shell echo ${BRANCH} | sed 's|/|_|g')
	@docker tag ${DOCKER_IMAGE_NAME}:${DOCKER_IMAGE_TAG} ${DOCKER_IMAGE_NAME}:latest
	@echo Successfully tagged ${DOCKER_IMAGE_NAME}:latest
	@echo "✅ Finished building Docker image!"

DATA_FOLDER="/data/fuelsequencer"
COMMAND?=""
run-docker-container:
	@echo "🤖 Running Docker image..."
	@docker run -d \
    		-v $(shell pwd)${DATA_FOLDER}:/home/fuelsequencer/.fuelsequencer \
    		--name $(DOCKER_CONTAINER_NAME) \
    		-p 26656:26656 -p 26657:26657 -p 1317:1317 \
    		${DOCKER_IMAGE_NAME}:latest \
    		${COMMAND}

start-docker-container:
	@echo "🤖 Starting Docker container..."
	@docker start $(DOCKER_CONTAINER_NAME)
	@echo "✅ Started Docker container!"

stop-docker-container:
	@echo "🤖 Stopping Docker container..."
	@docker stop $(DOCKER_CONTAINER_NAME)
	@echo "✅ Stopped Docker container!"

remove-docker-container:
	@echo "🤖 Removing Docker container..."
	@docker rm -v $(DOCKER_CONTAINER_NAME)
	@echo "✅ Removed Docker container!"

follow-docker-logs:
	@docker logs -f $(DOCKER_CONTAINER_NAME)

###############################################################################
###                                   E2E                                   ###
###############################################################################

check-ethereum-docker-image-exists:
ifeq (,$(shell docker images -q ${ETH_DOCKER_IMAGE_NAME}:latest 2> /dev/null))
	@echo "❌ Docker image ${ETH_DOCKER_IMAGE_NAME}:latest not found";
	@exit 1;
else
	@echo "✅ Found docker image ${ETH_DOCKER_IMAGE_NAME}:latest"
endif

build-ethereum-docker-image:
	@echo "🤖 Updating git submodules..."
	@git submodule init # for the first time
	@git submodule update --remote
	@# No need to add echos here, since `make build` has its own.
	@(cd e2e/test-contracts && make build)
	@echo "🤖 Cleaning up git submodules..."
	@git submodule update
	@echo "✅ Finished cleaning up git submodules!"

test-e2e-basic: check-docker-image-exists check-ethereum-docker-image-exists
	@cd e2e/tests && go test -mod=readonly -race -v ./basic/... --test.timeout 0

clean-e2e:
	@echo "🧹 Stopping Docker containers..."
	@docker ps -aq --filter "name=fuelsequencer0" | xargs -r docker stop
	@docker ps -aq --filter "name=fuelsequencer1" | xargs -r docker stop
	@docker ps -aq --filter "name=fuelsequencer2" | xargs -r docker stop
	@docker ps -aq --filter "name=ethereum" | xargs -r docker stop

	@echo "🧹 Removing Docker containers..."
	@docker ps -aq --filter "name=fuelsequencer0" | xargs -r docker rm
	@docker ps -aq --filter "name=fuelsequencer1" | xargs -r docker rm
	@docker ps -aq --filter "name=fuelsequencer2" | xargs -r docker rm
	@docker ps -aq --filter "name=ethereum" | xargs -r docker rm

	@echo "🧹 Pruning Docker networks..."
	@docker network prune -f

	@echo "✅ Finished cleaning E2E!"
