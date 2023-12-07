package hub

import "context"

// Hub интерфейс описывающий структуру данных, в которой хранятся все подключенные клиенты
type Hub[T any] interface {
	// AddClient добавляет клиента в хаб
	AddClient(ctx context.Context, c Client[T]) error
	// RemoveClient удаляет клиента из хаба
	RemoveClient(ctx context.Context, c Client[T]) error
	// Broadcast рассылает сообщение всем клиентам.
	// Можно передать BroadcastFilter для фильтрации тех клиентов,
	// кому должно быть доставлено сообщение
	Broadcast(ctx context.Context, msg T, filters ...BroadcastFilter) error
}

// BroadcastFilter служит для фильтрации клиентов при рассылке сообщений
type BroadcastFilter interface {
	filter(c uint) bool
}

// ExcludeClientFilter исключает клиента из рассылки
type ExcludeClientFilter struct {
	cid uint
}

func (e ExcludeClientFilter) filter(cid uint) bool {
	return e.cid != cid
}

func NewExcludeClientFilter(cid uint) *ExcludeClientFilter {
	return &ExcludeClientFilter{cid: cid}
}
