package server

import (
	"context"
	"fmt"
	"net"
	"p2p/codec"
	"p2p/log"
	"time"
)

type netClient[T any] struct {
	id    uint
	c     net.Conn
	codec codec.Codec[T]
}

func newNetClient[T any](id uint, c net.Conn, codec codec.Codec[T]) *netClient[T] {
	return &netClient[T]{id: id, c: c, codec: codec}
}

func (n *netClient[T]) Id() uint {
	return n.id
}

func (n *netClient[T]) Name() string {
	return fmt.Sprintf("Client #%d", n.id)
}

func (n *netClient[T]) Read(ctx context.Context) (msg T, err error) {
	var buff = make([]byte, 1024)
	size, err := n.c.Read(buff)
	if err != nil {
		return msg, err
	}

	var b = make([]byte, size)
	copy(b, buff[:size])

	decodedMsg, err := n.codec.Decode(b)
	if err != nil {
		log.Err("error decode msg. %w", err)
	}

	return decodedMsg, nil
}

func (n *netClient[T]) Send(ctx context.Context, msg T) error {
	b, err := n.codec.Encode(msg)
	if err != nil {
		return fmt.Errorf("error encode msg. %w", err)
	}

	if len(b) > 1024 {
		return fmt.Errorf("error message is too big. Max 1024 bytes")
	}

	err = n.c.SetWriteDeadline(time.Now().Add(time.Second))
	if err != nil {
		return fmt.Errorf("error set write deadline. %w", err)
	}
	_, err = n.c.Write(b)
	if err != nil {
		return fmt.Errorf("error writer msg. %w", err)
	}

	return nil
}
