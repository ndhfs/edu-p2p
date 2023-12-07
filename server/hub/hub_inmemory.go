package hub

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"
)

type InMemoryHub[T any] struct {
	mu      sync.RWMutex
	clients map[uint]Client[T]

	broadcastSem chan struct{}
}

func NewInMemoryHub[T any](cap int, broadcastThreads uint) *InMemoryHub[T] {
	if broadcastThreads == 0 {
		broadcastThreads = 1
	}

	return &InMemoryHub[T]{
		clients:      make(map[uint]Client[T], cap),
		broadcastSem: make(chan struct{}, broadcastThreads),
	}
}

func (i *InMemoryHub[T]) AddClient(ctx context.Context, c Client[T]) error {
	i.mu.Lock()
	defer i.mu.Unlock()

	i.clients[c.Id()] = c
	return nil
}

func (i *InMemoryHub[T]) RemoveClient(ctx context.Context, c Client[T]) error {
	i.mu.Lock()
	defer i.mu.Unlock()
	delete(i.clients, c.Id())
	return nil
}

func (i *InMemoryHub[T]) Broadcast(ctx context.Context, msg T, filters ...BroadcastFilter) error {
	select {
	case i.broadcastSem <- struct{}{}:
	case <-time.After(time.Second):
		return fmt.Errorf("broadcast service is busy now. Please, try later")
	}

	defer func() {
		<-i.broadcastSem
	}()

	clientsToSend := i.filterClients(filters)

	for _, c := range clientsToSend {
		if err := c.Send(ctx, msg); err != nil {
			log.Println("Error send msg to client ", c.Id(), ". Err: ", err)
		}
	}
	return nil
}

func (i *InMemoryHub[T]) filterClients(filters []BroadcastFilter) []Client[T] {
	i.mu.RLock()
	defer i.mu.RUnlock()
	out := make([]Client[T], 0, len(i.clients))
	var skip bool
	for _, c := range i.clients {
		skip = false
		for _, filter := range filters {
			if !filter.filter(c.Id()) {
				skip = true
				break
			}
		}
		if skip {
			continue
		}

		out = append(out, c)
	}
	return out
}
