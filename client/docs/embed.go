package docs

import "embed"

// REF: https://github.com/cosmos/cosmos-sdk/blob/v0.50.7/client/docs/embed.go

//go:embed swagger-ui
var SwaggerUI embed.FS
