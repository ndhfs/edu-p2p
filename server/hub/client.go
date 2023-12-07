package hub

import "context"

// Client описывает интерфейс клиента
type Client[T any] interface {
	Id() uint
	Name() string
	Send(ctx context.Context, msg T) error
}
