package client

import (
	"p2p/client"
	"p2p/server"
)

type Client[T any] struct {
	c *client.Client[T]
	s *server.TcpServer[T]
}
