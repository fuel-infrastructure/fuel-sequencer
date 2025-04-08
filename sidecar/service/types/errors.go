package types

import "fmt"

const (
	GeneratedRawTxBytesExceededMaxBytesStr = "generated raw tx bytes exceeded max bytes"
)

type GeneratedRawTxBytesExceededMaxBytesError struct {
	txSize   uint64
	maxBytes uint64
}

func (e *GeneratedRawTxBytesExceededMaxBytesError) Error() string {
	return fmt.Sprintf("%s; %d > %d", GeneratedRawTxBytesExceededMaxBytesStr, e.txSize, e.maxBytes)
}

func NewGeneratedRawTxBytesExceededMaxBytesError(txSize, maxBytes uint64) error {
	return &GeneratedRawTxBytesExceededMaxBytesError{txSize: txSize, maxBytes: maxBytes}
}
