package model

type WebResponse[T any] struct {
	Data   T      `json:"data,omitempty"`
	Errors string `json:"errors,omitempty"`
}
