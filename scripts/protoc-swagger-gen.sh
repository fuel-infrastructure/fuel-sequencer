#!/usr/bin/env bash

# REF: https://github.com/cosmos/cosmos-sdk/blob/v0.50.7/scripts/protoc-swagger-gen.sh
# But with `./cosmos` changed to `./fuelsequencer`
# TODO: add path of Cosmos SDK modules to include them in the output.

set -eo pipefail

mkdir -p ./tmp-swagger-gen
cd proto
proto_dirs=$(find ./fuelsequencer -path -prune -o -name '*.proto' -print0 | xargs -0 -n1 dirname | sort | uniq)
for dir in $proto_dirs; do
  # generate swagger files (filter query files)
  query_file=$(find "${dir}" -maxdepth 1 \( -name 'query.proto' -o -name 'service.proto' \))
  if [[ ! -z "$query_file" ]]; then
    buf generate --template buf.gen.swagger.yaml $query_file
  fi
done

cd ..
# combine swagger files
# uses nodejs package `swagger-combine`.
# all the individual swagger files need to be configured in `config.json` for merging
swagger-combine ./client/docs/config.json -o ./client/docs/swagger-ui/swagger.yaml -f yaml --continueOnConflictingPaths true --includeDefinitions true

# clean swagger files
rm -rf ./tmp-swagger-gen