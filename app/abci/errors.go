package abci

import (
	"fmt"

	sidecartypes "github.com/fuel-infrastructure/fuel-sequencer/sidecar/service/types"
)

const (
	FailedToEncodeEventAsRawTxBytesStr = "failed to encode event as raw tx bytes with err"
)

var _ error = &failedToEncodeEventAsRawTxBytesError{}

type failedToEncodeEventAsRawTxBytesError struct {
	err   error
	event *sidecartypes.Event
}

func NewFailedToEncodeEventAsRawTxBytesError(
	err error,
	event *sidecartypes.Event,
) *failedToEncodeEventAsRawTxBytesError {
	return &failedToEncodeEventAsRawTxBytesError{
		err:   err,
		event: event,
	}
}

func (e *failedToEncodeEventAsRawTxBytesError) Error() string {
	return fmt.Sprintf("%s: %s; event: %s", FailedToEncodeEventAsRawTxBytesStr, e.err.Error(), e.event)
}

func (e *failedToEncodeEventAsRawTxBytesError) LoggableKVs() []any {
	return []any{
		"reason", FailedToEncodeEventAsRawTxBytesStr,
		"error", e.Error(),
		"event", e.event,
	}
}
