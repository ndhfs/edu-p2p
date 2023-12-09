package main

import (
	"bufio"
	"context"
	"flag"
	"os"
	"os/signal"
	"p2p/client"
	"p2p/codec"
	"p2p/dto"
	"p2p/log"
	"syscall"
)

func main() {
	var serverAddr string
	flag.StringVar(&serverAddr, "server-addr", "127.0.0.1:8086", "server address to connect to")

	ctx, doneFn := signal.NotifyContext(
		context.Background(),
		syscall.SIGKILL, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT,
	)
	defer doneFn()

	jsonCodec := codec.NewJson[dto.Common]()
	c, err := client.Connect[dto.Common](serverAddr, jsonCodec)
	if err != nil {
		log.Fatal("error connect to server. %w", err)
	}
	defer c.Close()

	go func() {
		defer doneFn()
		c.Handle(ctx, client.HandlerFunc[dto.Common](func(ctx context.Context, data dto.Common) {
			log.Info("%s: %s", data.Message.From.Name, data.Message.Text)
		}))
	}()

	go func() {
		scanner := bufio.NewScanner(os.Stdin)
		for {
			if scanner.Scan() {
				msg := dto.NewTextMessage("", scanner.Text())
				log.Info("You: %s", msg.Message.Text)
				c.Send(context.Background(), dto.NewTextMessage("", scanner.Text()))
			}
		}
	}()

	<-ctx.Done()
}
