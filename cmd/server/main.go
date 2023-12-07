package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"p2p/codec"
	"p2p/dto"
	"p2p/log"
	"p2p/server"
	"p2p/server/hub"
)

func main() {
	// Адрес, на котором сервер будет ожидать соединения
	var listenAddr string

	// считаем флаги
	flag.StringVar(&listenAddr, "addr", ":8086", "tcp server listen on")
	flag.Parse()

	doneCtx, doneFn := signal.NotifyContext(context.Background(), os.Kill, os.Interrupt)
	defer func() {
		doneFn()
	}()

	clientsHub := hub.NewInMemoryHub[dto.Common](100, 20)
	jsonCodec := codec.NewJson[dto.Common]()

	srv := server.NewTcpServer[dto.Common](
		clientsHub,
		listenAddr,
		server.HandlerFunc[dto.Common](func(ctx context.Context, c hub.Client[dto.Common], msg dto.Common) {
			m := msg.Message
			log.Info("new message from client %d: %s", c.Id(), msg.Message.Text)
			m.From.Id = c.Id()
			m.From.Name = c.Name()

			if err := clientsHub.Broadcast(ctx, msg, hub.NewExcludeClientFilter(c.Id())); err != nil {
				log.Err("error broadcast message: %w")
			}
		}),
		jsonCodec,
	)

	var errCh chan error
	go func() {
		if err := srv.Serve(); err != nil {
			errCh <- fmt.Errorf("error serve tcp server. %w", err)
		}
	}()

	select {
	case err := <-errCh:
		log.Fatal("error start tcp server. %w", err)
	case <-doneCtx.Done():
		_ = srv.Shutdown()
		log.Info("shutdown server")
	}
}
