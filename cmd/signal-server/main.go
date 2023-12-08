package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	hub2 "p2p/cmd/signal-server/hub"
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

	// Создаем хаб для сигнального сервера
	clientsHub := hub2.NewSignalHub(
		hub.NewInMemoryHub[dto.SignalCommon](100, 20),
	)

	jsonCodec := codec.NewJson[dto.SignalCommon]()

	srv := server.NewTcpServer[dto.SignalCommon](
		clientsHub,
		listenAddr,
		// слушать нам нечего.
		server.HandlerFunc[dto.SignalCommon](func(ctx context.Context, c hub.Client[dto.SignalCommon], msg dto.SignalCommon) {
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
