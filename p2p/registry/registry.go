package registry

type Peer struct {
	Id   string
	Name string
	Addr string
}

func NewPeer(id string, name string, addr string) Peer {
	return Peer{Id: id, Name: name, Addr: addr}
}

type Registry interface {
	Register(peer Peer) error
	Unregister() error
	Peers() <-chan []Peer
}
