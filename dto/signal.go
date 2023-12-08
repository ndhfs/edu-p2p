package dto

type SignalCommon struct {
	Handshake *SignalHandshake `json:"handshake,omitempty"`
	Peers     *SignalPeers     `json:"peers,omitempty"`
}

type Peer struct {
	Name string `json:"name,omitempty"`
	Addr string `json:"addr,omitempty"`
}

type SignalHandshake struct {
	Peer
}

type SignalPeers struct {
	Peers []Peer `json:"peers,omitempty"`
}
