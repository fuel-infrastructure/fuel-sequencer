package types

import fmt "fmt"

const (
	AuthorizeEventTooManyMessagesStr       = "authorize event has too many messages"
	GeneratedRawTxBytesExceededMaxBytesStr = "generated raw tx bytes exceeded max bytes"
)

type authorizeEventTooManyMessagesError struct {
	actual   uint64
	expected uint64
}

func (e *authorizeEventTooManyMessagesError) Error() string {
	return fmt.Sprintf("%s; %d > %d", AuthorizeEventTooManyMessagesStr, e.actual, e.expected)
}

func AuthorizeEventTooManyMessagesError(actual, expected uint64) error {
	return &authorizeEventTooManyMessagesError{actual: actual, expected: expected}
}

type generatedRawTxBytesExceededMaxBytesError struct {
	txSize   uint64
	maxBytes uint64
}

func (e *generatedRawTxBytesExceededMaxBytesError) Error() string {
	return fmt.Sprintf("%s; %d > %d", GeneratedRawTxBytesExceededMaxBytesStr, e.txSize, e.maxBytes)
}

func GeneratedRawTxBytesExceededMaxBytesError(txSize, maxBytes uint64) error {
	return &generatedRawTxBytesExceededMaxBytesError{txSize: txSize, maxBytes: maxBytes}
}
