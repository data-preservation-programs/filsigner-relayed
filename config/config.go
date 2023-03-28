package config

import (
	"github.com/ipfs/go-log/v2"
	"github.com/libp2p/go-libp2p/core/peer"
)

func GetDefaultRelayInfo() []peer.AddrInfo {
	relays := make([]peer.AddrInfo, 2)

	naRelayInfo, err := peer.AddrInfoFromString("/dns4/relay-na.spade.services/tcp/4001/p2p/12D3KooWBVheEM7TdvfQHNLsGy39PFuDSXnnkHyXfgH5uD1pheqv")
	if err != nil {
		log.Logger("config").Fatalf("failed to parse naRelayInfo: %v", err)
	}

	euRelayInfo, err := peer.AddrInfoFromString("/dns4/relay-eu.spade.services/tcp/4001/p2p/12D3KooWGxyLaT4h4XYYrcCpRVHh5N3WNTLJmCtaKHrfVz7sfTjM")
	if err != nil {
		log.Logger("config").Fatalf("failed to parse euRelayInfo: %v", err)
	}

	relays[0] = *naRelayInfo
	relays[1] = *euRelayInfo
	return relays
}

const ProtocolName = "/cmd/signproposal/poc"
