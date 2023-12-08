package mdns

import (
	"fmt"
	"github.com/hashicorp/mdns"
	"net"
	"p2p/log"
	"p2p/p2p/registry"
	"strconv"
	"strings"
)

type Registry struct {
	server  *mdns.Server
	curPeer registry.Peer
}

func NewRegistry() *Registry {
	return &Registry{}
}

func (r *Registry) Peers() <-chan []registry.Peer {
	peerCh := make(chan []registry.Peer, 1)
	go func() {
		for {
			// Make a channel for results and start listening
			entriesCh := make(chan *mdns.ServiceEntry, 4)
			go func() {
				var peers []registry.Peer
				for entry := range entriesCh {
					if strings.Contains(entry.Name, "lab2.p2p.demo.") {
						peer := registry.Peer{
							Id:   entry.InfoFields[0],
							Name: entry.InfoFields[1],
							Addr: entry.InfoFields[2],
						}
						if peer.Name == "" {
							peer.Name = peer.Addr
						}
						if r.curPeer.Id != peer.Id {
							peers = append(peers, peer)
						}
					}
				}
				peerCh <- peers
			}()

			// Start the lookup
			if err := mdns.Query(&mdns.QueryParam{
				Service:     "lab2",
				Entries:     entriesCh,
				Domain:      "p2p.demo.",
				DisableIPv6: true,
			}); err != nil {
				log.Err("error lookup service. %w", err)
				return
			}

			close(entriesCh)
		}
	}()
	return peerCh
}

func (r *Registry) Register(peer registry.Peer) error {
	r.curPeer = peer
	host, portStr, err := net.SplitHostPort(peer.Addr)
	if err != nil {
		return fmt.Errorf("error parse node address format '%s'. %w", peer.Addr, err)
	}

	port, err := strconv.Atoi(portStr)
	if err != nil {
		return fmt.Errorf("error parse node address port '%s'. %w", portStr, err)
	}

	info := []string{
		peer.Id,
		peer.Name,
		peer.Addr,
	}

	// we got here, new node
	s, err := mdns.NewMDNSService(
		peer.Id,
		"lab2",
		"p2p.demo.",
		"",
		port,
		[]net.IP{net.ParseIP(host)},
		info,
	)
	if err != nil {
		return fmt.Errorf("error create mdns instance. %w", err)
	}

	r.server, err = mdns.NewServer(&mdns.Config{Zone: s})
	if err != nil {
		return fmt.Errorf("error create mdns server. %w", err)
	}

	return nil
}

func (r *Registry) Unregister() error {
	return r.server.Shutdown()
}
