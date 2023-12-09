package main

import (
	"bufio"
	"context"
	"encoding/base64"
	"flag"
	"fmt"
	"github.com/google/uuid"
	"os"
	"os/signal"
	"p2p/client"
	"p2p/codec"
	"p2p/dto"
	"p2p/log"
	"p2p/p2p/registry"
	"p2p/p2p/registry/mdns"
	signal2 "p2p/p2p/registry/signal"
	"p2p/server"
	"p2p/server/hub"
	"path/filepath"
	"strings"
	"time"
)

func main() {

	// Дадим возможность указать свое имя
	var peerName string
	var registryType string
	var signalAddr string

	flag.StringVar(&peerName, "name", "", "set peer name")
	flag.StringVar(&registryType, "registry", "mdns", "set registry type: mdns or signal")
	flag.StringVar(&signalAddr, "signal-addr", "127.0.0.1:8086", "set signal server address")
	flag.Parse()

	// Ждем сигнала о завершении процесса
	doneCtx, doneFn := signal.NotifyContext(context.Background(), os.Kill, os.Interrupt)
	defer func() {
		doneFn()
	}()

	// Создаем канал, куда будем передавать входящие сообщения
	var incomingMessageCh = make(chan *dto.Message)

	// Создаем объект хаба для сервера
	memoryHub := hub.NewInMemoryHub[dto.Common](100, 20)
	// Создаем объект кодека Json, который будет преобразовывать байты в структуру и обратно
	jsonCodec := codec.NewJson[dto.Common]()
	// Создаем объект сервера
	srv := server.NewTcpServer[dto.Common](
		memoryHub,
		"127.0.0.1:0",
		server.HandlerFunc[dto.Common](func(ctx context.Context, c hub.Client[dto.Common], msg dto.Common) {
			if msg.Message.From.Name == "" {
				msg.Message.From.Name = c.Name()
			}
			incomingMessageCh <- msg.Message
		}),
		codec.NewJson[dto.Common](),
	)

	// Начинаем слушать сокет.
	var errCh chan error
	go func() {
		if err := srv.Serve(); err != nil {
			errCh <- fmt.Errorf("error serve tcp server. %w", err)
		}
	}()
	select {
	case err := <-errCh:
		log.Fatal("error start tcp server. %w", err)
	case <-time.After(100 * time.Millisecond):
	}

	// Создаем экземпляр реестра, мультикаст, чтобы сообщить о себе и найти другие пиры в сети
	var reg registry.Registry
	switch registryType {
	case "mdns":
		reg = mdns.NewRegistry()
	case "signal":
		reg = signal2.NewRegistry(signalAddr)
	}
	// Определим адрес, на котором мы слушаем сокет
	addr := fmt.Sprintf("127.0.0.1:%d", srv.Port())
	// Заявляем о себе в сети
	if err := reg.Register(registry.NewPeer(uuid.New().String(), peerName, addr)); err != nil {
		log.Fatal("error register peer. %w", err)
	}
	// Когда процесс завершится, мы перестанем отправлять пакеты в мультикаст
	defer func() {
		_ = reg.Unregister()
	}()

	// Получаем от реестра канал, куда будет передавать актуальный список пиров
	peerWatchCh := reg.Peers()

	// Ожидаем ввод пользователя и передаем в канал cmdCh
	var cmdCh = make(chan string)
	go func() {
		scanner := bufio.NewScanner(os.Stdin)
		for {
			if scanner.Scan() {
				cmdCh <- scanner.Text()
			}
		}
	}()

	// В эту мапу мы будем складывать пиры по имени для более быстрого поиска
	var availablePeers map[string]registry.Peer

	// Текущий пир, к которому мы подключились
	var selectedPeer *registry.Peer
	// Соединение с текущим пиром
	var selectedPeerConn *client.Client[dto.Common]

	// Собственно точка синхронизации, где мы в цикле получаем сообщения их разных каналов
	for {
		select {
		// Получили обновление списка пиров
		case peers := <-peerWatchCh:
			availablePeers = make(map[string]registry.Peer)
			for _, peer := range peers {
				availablePeers[peer.Name] = peer
			}
		// Получили новую команду от клиента
		case cmd := <-cmdCh:
			args := strings.Split(cmd, " ")
			switch args[0] {
			// Команда на получение списка пиров
			case "/list":
				for _, peer := range availablePeers {
					fmt.Printf("@%s\n", peer.Name)
				}

			// Команда на подключение к определенному пиру по имени
			case "/switch":
				peerName := strings.TrimLeft(args[1], "@")
				// Ищем пир в нашей мапе
				peer, ok := availablePeers[peerName]
				// Если не найдем - выводим ошибку
				if !ok {
					log.Err("Peer %s not found.", peerName)
					break
				}

				if selectedPeer != nil {
					if selectedPeer.Name == peerName {
						log.Info("already connected to %s", peerName)
						break
					}

					log.Info("disconnect from %s", selectedPeer.Name)
					selectedPeerConn.Close()
					selectedPeer = nil
				}

				// Пытаемся подключиться к пиру
				log.Info("connecting to peer %s on %s", peer.Name, peer.Addr)
				peerConn, err := client.Connect[dto.Common](peer.Addr, jsonCodec)
				if err != nil {
					log.Err("error connect to peer %s on %s. %w", peer.Name, peer.Addr, err)
					break
				}

				selectedPeerConn = peerConn
				selectedPeer = &peer
				log.Info("connected")
			case "/file":
				filePath := args[1]
				mediaMessage, err := dto.NewMediaMessage(peerName, filePath)
				if err != nil {
					log.Err("error create message with file. %w", err)
				}
				if err := selectedPeerConn.Send(context.Background(), mediaMessage); err != nil {
					log.Err("error send message to current peer. %w", err)
				}
				fmt.Printf("@You > @%s: [File: %s]\n", selectedPeer.Name, filePath)
			default:
				if selectedPeerConn == nil {
					log.Err("No peer connected. Use /switch [@PeerName] to connect.")
					break
				}

				if err := selectedPeerConn.Send(context.Background(), dto.NewTextMessage(peerName, cmd)); err != nil {
					log.Err("error send message to current peer. %w", err)
					break
				}
				fmt.Printf("@You > @%s: %s\n", selectedPeer.Name, cmd)
			}
		// Входящее сообщение
		case inMsg := <-incomingMessageCh:
			if inMsg.Media != nil {
				storagePath := "storage/" + peerName
				if _, err := os.Stat(storagePath); os.IsNotExist(err) {
					if err := os.MkdirAll("storage/"+peerName, 0775); err != nil {
						log.Err("error create storage dir. %w", err)
						break
					}
				}

				fileContent, err := base64.StdEncoding.DecodeString(inMsg.Media.Content)
				if err != nil {
					log.Err("error decode file content. %w", err)
					break
				}

				filePath, _ := filepath.Abs(storagePath + "/" + inMsg.Media.Filename)
				if err := os.WriteFile(filePath, fileContent, 0775); err != nil {
					log.Err("error save file. %w", err)
					break
				}

				fmt.Printf("@%s > @You: [File %s]\n", inMsg.From.Name, filePath)
			} else {
				fmt.Printf("@%s > @You: %s\n", inMsg.From.Name, inMsg.Text)
			}
		case <-doneCtx.Done():
			_ = srv.Shutdown()
			log.Info("shutdown server")
			return
		}
	}
}
