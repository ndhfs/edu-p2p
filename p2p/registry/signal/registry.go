package signal

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"p2p/client"
	"p2p/codec"
	"p2p/dto"
	"p2p/p2p/registry"
)

type Registry struct {
	Addr    string
	c       *client.Client[dto.SignalCommon]
	peersCh chan []registry.Peer
}

func (r *Registry) Register(peer registry.Peer) error {
	c, err := client.Connect[dto.SignalCommon](r.Addr, codec.NewJson[dto.SignalCommon]())
	if err != nil {
		return fmt.Errorf("error connect to signal server. %w", err)
	}
	go c.Handle(context.Background(), client.HandlerFunc[dto.SignalCommon](func(ctx context.Context, data dto.SignalCommon) {
		if data.Peers != nil {
			var peers = make([]registry.Peer, 0, len(data.Peers.Peers))
			for _, p := range data.Peers.Peers {
				if p.Name != peer.Name {
					peers = append(peers, registry.NewPeer(uuid.New().String(), p.Name, p.Addr))
				}
			}
			r.peersCh <- peers
		}
	}))
	if err := r.handshake(c, peer); err != nil {
		return fmt.Errorf("error handshake with signal. %w", err)
	}
	r.c = c
	return nil
}

func (r *Registry) Unregister() error {
	r.c.Close()
	return nil
}

func (r *Registry) Peers() <-chan []registry.Peer {
	return r.peersCh
}

func (r *Registry) handshake(c *client.Client[dto.SignalCommon], peer registry.Peer) error {
	err := c.Send(context.Background(), dto.SignalCommon{
		Handshake: &dto.SignalHandshake{
			Peer: dto.Peer{
				Name: peer.Name,
				Addr: peer.Addr,
			},
		},
	})
	if err != nil {
		return fmt.Errorf("error send handshake package. %w", err)
	}

	return nil
}

func NewRegistry(addr string) *Registry {
	return &Registry{Addr: addr, peersCh: make(chan []registry.Peer, 10)}
}
