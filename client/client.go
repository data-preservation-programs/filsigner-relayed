package client

import (
	"context"
	"github.com/data-preservation-programs/filsigner-relayed/model"
	cborutil "github.com/filecoin-project/go-cbor-util"
	filmarket "github.com/filecoin-project/go-state-types/builtin/v9/market"
	filcrypto "github.com/filecoin-project/go-state-types/crypto"
	"github.com/jsign/go-filsigner/wallet"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/peerstore"
	ma "github.com/multiformats/go-multiaddr"
	"github.com/pkg/errors"
	"time"
)

type Client struct {
	host   host.Host
	relays []peer.AddrInfo
}

func (c Client) SignProposal(ctx context.Context, dest peer.ID, proposal filmarket.DealProposal) (*filcrypto.Signature, error) {
	targetAddrs := make([]ma.Multiaddr, 0)
	for _, relay := range c.relays {
		for _, addr := range relay.Addrs {
			targetAddr, err := ma.NewMultiaddr(addr.String() + "/p2p/" + relay.ID.String() + "/p2p-circuit/p2p/" + dest.String())
			if err != nil {
				return nil, errors.Wrap(err, "failed to create target relayed multiaddr")
			}

			targetAddrs = append(targetAddrs, targetAddr)
		}
	}

	c.host.Peerstore().AddAddrs(dest, targetAddrs, peerstore.PermanentAddrTTL)

	// Marshal and send out the proposal
	proposalBytes, err := cborutil.Dump(&proposal)
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshall proposal")
	}

	stream, err := c.host.NewStream(network.WithUseTransient(ctx, "signproposal"), dest, "/filsigner-relayed/signproposal/v1")
	if err != nil {
		return nil, errors.Wrap(err, "failed to open stream")
	}

	defer stream.Close()
	if deadline, ok := ctx.Deadline(); ok {
		stream.SetDeadline(deadline)
		defer stream.SetDeadline(time.Time{})
	}

	_, err = stream.Write(proposalBytes)
	if err != nil {
		return nil, errors.Wrap(err, "failed to write proposal to stream")
	}
	stream.CloseWrite()

	response := new(model.SignerResponse)
	err = cborutil.ReadCborRPC(stream, response)
	if err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal response")
	}

	if response.Code != model.Success {
		return nil, &RequestError{
			StatusCode: response.Code,
			Message:    response.Message,
		}
	}

	signature := new(filcrypto.Signature)

	// Verify the signature
	valid, err := wallet.WalletVerify(proposal.Client, proposalBytes, response.Signature)
	if err != nil {
		return nil, errors.Wrap(err, "failed to verify signature")
	}

	if !valid {
		return nil, errors.New("signature is not valid")
	}

	err = signature.UnmarshalBinary(response.Signature)
	if err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal response signature")
	}
	return signature, nil
}

// NewClient creates a new client with the default relays
// @param privateKey the private key to use for the libp2p host
func NewClient(privateKey crypto.PrivKey, relays []peer.AddrInfo) (*Client, error) {
	host, err := libp2p.New(
		libp2p.NoListenAddrs,
		libp2p.EnableRelay(),
		libp2p.Identity(privateKey),
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create libp2p host")
	}

	return NewClientWithHost(host, relays)
}

// NewClientWithHost creates a new client with the default relays
// @param libp2p the libp2p host. This libp2p instance must have Relay enabled
func NewClientWithHost(host host.Host, relays []peer.AddrInfo) (*Client, error) {
	client := &Client{
		host:   host,
		relays: relays,
	}

	return client, nil
}
