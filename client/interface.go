package client

import (
	"context"
	filmarket "github.com/filecoin-project/go-state-types/builtin/v9/market"
	"github.com/filecoin-project/go-state-types/crypto"
	"github.com/libp2p/go-libp2p/core/peer"
)

type Signer interface {
	SignProposal(ctx context.Context, dest peer.ID, proposal filmarket.DealProposal) (*crypto.Signature, error)
}
