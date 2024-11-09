# Generate protobuf files for rust api client

set -eo pipefail

(rm -f rust_client/protos/*.rs)

# Generating rust API module
(buf generate --include-imports --verbose --template proto/buf.gen.rust.yaml)
