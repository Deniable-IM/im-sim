package network

import (
	"deniable-im/im-sim/internal/types"
)

type AddrMapping struct {
	pair types.Pair[*Network, string]
}

func NewAddrMapping(network *Network, ipv4 string) AddrMapping {
	return AddrMapping{types.Pair[*Network, string]{Fst: network, Snd: ipv4}}
}

func (addrMapping AddrMapping) GetConnection() *Connection {
	return &Connection{Network: addrMapping.pair.Fst, IPv4: &addrMapping.pair.Snd}
}
