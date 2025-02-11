#!/usr/bin/env bash

# The below steps are for generating mocks that are useful for testing
# Ref: https://github.com/cosmos/cosmos-sdk/blob/v0.50.3/scripts/mockgen.sh

mockgen_cmd="mockgen"
$mockgen_cmd -source=x/bridge/types/expected_keepers.go -package testutil -destination x/bridge/testutil/expected_keepers_mocks.go
$mockgen_cmd -source=x/sequencing/types/expected_keepers.go -package testutil -destination x/sequencing/testutil/expected_keepers_mocks.go
$mockgen_cmd -source=x/mint/types/expected_keepers.go -package testutil -destination x/mint/testutil/expected_keepers_mocks.go
$mockgen_cmd -source=x/reports/types/expected_keepers.go -package testutil -destination x/reports/testutil/expected_keepers_mocks.go
$mockgen_cmd -source=sidecar/client/interface.go -package testutil -destination sidecar/testutil/app_sidecar_client_mocks.go



