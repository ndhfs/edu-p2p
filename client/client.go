package client

import (
	"context"
	"fmt"
	"net"
	"p2p/codec"
	"p2p/log"
	"sync/atomic"
)

type Handler[T any] interface {
	Handle(ctx context.Context, data T)
}

type HandlerFunc[T any] func(ctx context.Context, data T)

func (h HandlerFunc[T]) Handle(ctx context.Context, data T) {
	h(ctx, data)
}

type Client[T any] struct {
	conn      net.Conn
	isClosing uint32
	h         Handler[T]
	codec     codec.Codec[T]
}

func Connect[T any](addr string, codec codec.Codec[T]) (*Client[T], error) {
	c := &Client[T]{
		codec: codec,
	}

	conn, err := net.Dial("tcp4", addr)
	if err != nil {
		return nil, fmt.Errorf("error dial connect with addr %s. %w", addr, err)
	}
	c.conn = conn
	return c, nil
}

func (c *Client[T]) Handle(ctx context.Context, handler Handler[T]) {
	defer func() {
		c.close()
	}()
	var buff = make([]byte, 1024)
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		if atomic.LoadUint32(&c.isClosing) == 1 {
			return
		}

		n, err := c.conn.Read(buff)
		if err != nil {
			log.Err("error read data from server. %w", err)
			return
		}

		msg := make([]byte, n)
		copy(msg, buff[:n])

		decodedMsg, err := c.codec.Decode(msg)
		if err != nil {
			log.Err("error decode msg. %w", err)
		}

		handler.Handle(context.Background(), decodedMsg)
	}
}

func (c *Client[T]) Send(ctx context.Context, msg T) error {
	b, err := c.codec.Encode(msg)
	if err != nil {
		log.Err("error encode msg. %w", err)
	}
	_, err = c.conn.Write(b)
	if err != nil {
		return fmt.Errorf("error write to conn. %w", err)
	}
	return nil
}

func (c *Client[T]) Close() {
	c.close()
}

func (c *Client[T]) close() {
	if atomic.CompareAndSwapUint32(&c.isClosing, 0, 1) {
		c.conn.Close()
	}
}
