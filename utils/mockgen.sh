#!/usr/bin/env bash

mockgen_cmd="mockgen"
$mockgen_cmd -source=x/bridge/types/expected_keepers.go -package testutil -destination x/bridge/testutil/expected_keepers_mocks.go
$mockgen_cmd -source=x/sequencing/types/expected_keepers.go -package testutil -destination x/sequencing/testutil/expected_keepers_mocks.go
