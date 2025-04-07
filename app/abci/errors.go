package abci

import (
	"fmt"

	sidecartypes "github.com/fuel-infrastructure/fuel-sequencer/sidecar/service/types"
)

const (
	FailedToEncodeEventAsRawTxBytesStr = "failed to encode event as raw tx bytes with err"
)

var _ error = &FailedToEncodeEventAsRawTxBytesError{}

type FailedToEncodeEventAsRawTxBytesError struct {
	err   error
	event *sidecartypes.Event
}

func NewFailedToEncodeEventAsRawTxBytesError(
	err error,
	event *sidecartypes.Event,
) *FailedToEncodeEventAsRawTxBytesError {
	return &FailedToEncodeEventAsRawTxBytesError{
		err:   err,
		event: event,
	}
}

func (e *FailedToEncodeEventAsRawTxBytesError) Error() string {
	return fmt.Sprintf("%s: %s", FailedToEncodeEventAsRawTxBytesStr, e.err.Error())
}

func (e *FailedToEncodeEventAsRawTxBytesError) LoggableKVs() []any {
	return []any{
		"reason", FailedToEncodeEventAsRawTxBytesStr,
		"error", e.Error(),
		"event", e.event,
	}
}
