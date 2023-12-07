package server

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"p2p/codec"
	"p2p/log"
	"p2p/server/hub"
	"sync/atomic"
)

type Handler[T any] interface {
	Handle(ctx context.Context, c hub.Client[T], msg T)
}

type HandlerFunc[T any] func(ctx context.Context, c hub.Client[T], msg T)

func (h HandlerFunc[T]) Handle(ctx context.Context, c hub.Client[T], msg T) {
	h(ctx, c, msg)
}

type TcpServer[T any] struct {
	hub         hub.Hub[T]
	addr        string
	handler     Handler[T]
	codec       codec.Codec[T]
	closingFlag uint32

	l net.Listener
}

func NewTcpServer[T any](
	hub hub.Hub[T],
	addr string,
	handler Handler[T],
	codec codec.Codec[T],
) *TcpServer[T] {
	return &TcpServer[T]{
		hub:     hub,
		addr:    addr,
		handler: handler,
		codec:   codec,
	}
}

func (s *TcpServer[T]) Serve() error {
	log.Info("Starting server on %s", s.addr)

	listener, err := net.Listen("tcp4", s.addr)
	if err != nil {
		return fmt.Errorf("error start listener. %w", err)
	}
	s.l = listener

	var connNum uint

	for {
		c, err := listener.Accept()
		if err != nil {
			if atomic.LoadUint32(&s.closingFlag) == 1 {
				return nil
			}

			log.Err("error accept next conn. %w", err)
			continue
		}
		connNum++
		conn := newNetClient[T](connNum, c, s.codec)
		err = s.hub.AddClient(context.Background(), conn)
		if err != nil {
			log.Err("error add client to hub. %w", err)
			_ = conn.c.Close()
		}

		go s.handleConn(conn)
	}
}

func (s *TcpServer[T]) Shutdown() error {
	atomic.CompareAndSwapUint32(&s.closingFlag, 0, 1)
	log.Info("Shutting down server on")
	return s.l.Close()
}

func (s *TcpServer[T]) handleConn(conn *netClient[T]) {
	log.Info("Client connected: %s", conn.Name())
	defer func() {
		_ = conn.c.Close()
		s.hub.RemoveClient(context.Background(), conn)
		log.Info("Client disconnected: %s", conn.Name())
	}()

	var buff = make([]byte, 1024)
	for {
		n, err := conn.c.Read(buff)
		if err != nil {
			if !errors.Is(err, io.EOF) {
				log.Err("error read msg from conn. %s, %w", conn.Id(), err)
			}
			return
		}

		var msg = make(hub.Msg, n)
		copy(msg, buff[:n])

		decodedMsg, err := s.codec.Decode(msg)
		if err != nil {
			log.Err("error decode msg. %w", err)
		}

		s.handler.Handle(context.Background(), conn, decodedMsg)
	}
}
