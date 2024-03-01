# The below steps are for generating protobuf files for the new google.golang.org/protobuf API
# Ref: https://github.com/cosmos/cosmos-sdk/blob/v0.50.3/scripts/protocgen-pulsar.sh

set -eo pipefail

# Cleaning API directory
(cd api; find ./ -type f \( -iname \*.pulsar.go -o -iname \*.pb.go -o -iname \*.cosmos_orm.go -o -iname \*.pb.gw.go \) -delete; find . -empty -type d -delete; cd ..)

# Generating API module
(buf generate --template proto/buf.gen.pulsar.yaml)
