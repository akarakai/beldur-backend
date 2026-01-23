package dto

// ListResponse is used as a wrapper for http JSON responses
type ListResponse[T any] struct {
	Data []T `json:"data"`
}
