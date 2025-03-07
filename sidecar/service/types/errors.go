package types

import "fmt"

const (
	AuthorizeEventTooManyMessagesStr       = "authorize event has too many messages"
	GeneratedRawTxBytesExceededMaxBytesStr = "generated raw tx bytes exceeded max bytes"
)

type AuthorizeEventTooManyMessagesError struct {
	actual   uint64
	expected uint64
}

func (e *AuthorizeEventTooManyMessagesError) Error() string {
	return fmt.Sprintf("%s; %d > %d", AuthorizeEventTooManyMessagesStr, e.actual, e.expected)
}

func NewAuthorizeEventTooManyMessagesError(actual, expected uint64) error {
	return &AuthorizeEventTooManyMessagesError{actual: actual, expected: expected}
}

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
