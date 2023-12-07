package codec

import "encoding/json"

type Json[T any] struct {
}

func NewJson[T any]() *Json[T] {
	return &Json[T]{}
}

func (j *Json[T]) Encode(t T) ([]byte, error) {
	return json.Marshal(t)
}

func (j *Json[T]) Decode(bytes []byte) (T, error) {
	var v T
	err := json.Unmarshal(bytes, &v)
	return v, err
}
