package hub

import (
	"context"
	"fmt"
	"p2p/dto"
	"p2p/server/hub"
	"time"
)

type SignalHub struct {
	hub.Hub[dto.SignalCommon]
	peers map[uint]dto.Peer
}

func NewSignalHub(hub hub.Hub[dto.SignalCommon]) *SignalHub {
	return &SignalHub{Hub: hub, peers: make(map[uint]dto.Peer)}
}

func (s *SignalHub) AddClient(ctx context.Context, c hub.Client[dto.SignalCommon]) error {
	// Перед тем как добавить клиента в хаб, нужно совершить рукопожатие. Узнать о нем информацию
	cc, err := s.handshake(c)
	if err != nil {
		return fmt.Errorf("error handshake. %w", err)
	}

	if err := s.Hub.AddClient(ctx, cc); err != nil {
		return err
	}

	s.peers[cc.Id()] = cc.peer
	s.notifyPeers()

	return nil
}

func (s *SignalHub) RemoveClient(ctx context.Context, c hub.Client[dto.SignalCommon]) error {
	if err := s.Hub.RemoveClient(ctx, c); err != nil {
		return err
	}

	delete(s.peers, c.Id())
	s.notifyPeers()

	return nil
}

func (s *SignalHub) handshake(c hub.Client[dto.SignalCommon]) (*NamedClient, error) {
	// На рукопожатие отводим не более 5 сек
	ctx, doneFn := context.WithTimeout(context.Background(), 5*time.Second)
	defer func() {
		doneFn()
	}()

	m, err := c.Read(ctx)
	if err != nil {
		return nil, fmt.Errorf("error read handshake package. %w", err)
	}

	if m.Handshake == nil {
		return nil, fmt.Errorf("error invalid handshake req. %w", err)
	}

	return NewNamedClient(c, m.Handshake.Name, m.Handshake.Addr), nil
}

func (s *SignalHub) notifyPeers() {
	var peers []dto.Peer
	for _, peer := range s.peers {
		peers = append(peers, peer)
	}

	s.Broadcast(context.Background(), dto.SignalCommon{
		Peers: &dto.SignalPeers{
			Peers: peers,
		},
	})
}

type NamedClient struct {
	hub.Client[dto.SignalCommon]
	peer dto.Peer
}

func (n *NamedClient) Name() string {
	return n.peer.Name
}

func (n *NamedClient) Addr() string {
	return n.peer.Addr
}

func NewNamedClient(client hub.Client[dto.SignalCommon], name string, addr string) *NamedClient {
	return &NamedClient{Client: client, peer: dto.Peer{Name: name, Addr: addr}}
}
