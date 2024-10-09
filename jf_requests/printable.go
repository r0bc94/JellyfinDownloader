package jf_requests

type Printable[T any] interface {
	PrintAndGetSelection() (T, error)
	PrintAndGetConfirmation() bool
}
