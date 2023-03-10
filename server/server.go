package server

import (
	"bytes"
	"context"
	"github.com/data-preservation-programs/filsigner-relayed/model"
	"github.com/filecoin-project/go-address"
	cborutil "github.com/filecoin-project/go-cbor-util"
	filmarket "github.com/filecoin-project/go-state-types/builtin/v9/market"
	cbornode "github.com/ipfs/go-ipld-cbor"
	logging "github.com/ipfs/go-log/v2"
	"github.com/jpillora/backoff"
	"github.com/jsign/go-filsigner/wallet"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/p2p/protocol/circuitv2/client"
	"github.com/pkg/errors"
	"io"
	"time"
)

type WalletPrivateKey = string

type Server struct {
	host              host.Host
	relays            []peer.AddrInfo
	allowedRequesters []peer.ID
	keyMap            map[address.Address]WalletPrivateKey
}

func NewServer(privateKey crypto.PrivKey, allowedRequesters []peer.ID, walletKeys []WalletPrivateKey, relays []peer.AddrInfo) (*Server, error) {
	keyMap := make(map[address.Address]WalletPrivateKey)
	for _, key := range walletKeys {
		addr, err := wallet.PublicKey(key)
		if err != nil {
			return nil, errors.Wrap(err, "failed to resolve private key to public key (address)")
		}

		keyMap[addr] = key
	}

	host, err := libp2p.New(
		libp2p.NoListenAddrs,
		libp2p.EnableRelay(),
		libp2p.Identity(privateKey),
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create libp2p host")
	}

	return &Server{
		host:              host,
		relays:            relays,
		allowedRequesters: allowedRequesters,
		keyMap:            keyMap,
	}, nil
}

func SendError(stream network.Stream, code model.StatusCode, message string) {
	log := logging.Logger("server")
	log.Errorw("sending error", "code", code, "message", message)
	err := cborutil.WriteCborRPC(stream, &model.SignerResponse{
		Code:    code,
		Message: message,
	})
	if err != nil {
		log.Errorw("failed to send error", "error", err)
	}
}

func (s Server) Start(ctx context.Context) error {
	log := logging.Logger("server")
	// Setup stream handler
	s.host.SetStreamHandler("/cmd/signproposal/v1", func(stream network.Stream) {
		log := log.With("remote", stream.Conn().RemotePeer().String())
		log.Info("got sign proposal request")
		defer stream.Close()

		// Verify that the request is from allowed requesters
		allowed := false
		for _, allowedRequester := range s.allowedRequesters {
			if stream.Conn().RemotePeer() == allowedRequester {
				allowed = true
				break
			}
		}
		if !allowed {
			SendError(stream, model.UnauthorizedRequester, "request is not from allowed requesters")
			return
		}

		// Read the proposal bytes
		request, err := io.ReadAll(stream)
		if err != nil {
			SendError(stream, model.ReadStreamError, err.Error())
			return
		}

		// Unmarshall to the proposal object
		proposal := new(filmarket.DealProposal)
		err = cbornode.DecodeInto(request, proposal)
		if err != nil {
			SendError(stream, model.DecodeRequestError, err.Error())
			return
		}

		log.Infow("proposal decoded", "proposal", proposal)

		// Verify the original proposal is properly marshalled
		proposalBytes, err := cborutil.Dump(proposal)
		if err != nil {
			SendError(stream, model.EncodeRequestError, err.Error())
			return
		}

		if !bytes.Equal(request, proposalBytes) {
			SendError(stream, model.ProposalRemarshalMismatch, "proposal remarshalled does not match the original proposal bytes")
			return
		}

		// Sign the proposal
		privateKey, ok := s.keyMap[proposal.Client]
		if !ok {
			SendError(stream, model.WalletKeyNotFound, "private key not found for the proposal client address "+proposal.Client.String())
			return
		}

		signature, err := wallet.WalletSign(privateKey, proposalBytes)
		if err != nil {
			SendError(stream, model.WalletSignError, err.Error())
			return
		}

		// Marshall the signature
		signatureBytes, err := signature.MarshalBinary()
		if err != nil {
			SendError(stream, model.MarshalSignatureError, err.Error())
		}

		response := &model.SignerResponse{
			Code:      model.Success,
			Signature: signatureBytes,
		}

		// Marshall the response
		responseBytes, err := cborutil.Dump(response)
		if err != nil {
			SendError(stream, model.EncodeResponseError, err.Error())
			return
		}

		// Send back the signature
		_, err = stream.Write(responseBytes)
		if err != nil {
			log.Errorw("failed to sent the response back", "error", err)
			return
		}
	})

	// Start connection to relay servers
	for _, relay := range s.relays {
		relay := relay
		log := log.With("relay", relay.ID.String())
		go func() {
			waitTime := &backoff.Backoff{
				Min: 10 * time.Second,
				Max: time.Minute,
			}

			for {
				disconnected := s.host.Network().Connectedness(relay.ID) == network.NotConnected
				if disconnected {
					log.Info("connecting to relay server")
					err := s.host.Connect(ctx, relay)
					if err != nil {
						log.Errorw("failed to connect to relay server", "error", err)
						select {
						case <-ctx.Done():
							return
						case <-time.After(waitTime.Duration()):
							continue
						}
					}

					log.Info("making reservation")
					reservation, err := client.Reserve(ctx, s.host, relay)
					if err != nil {
						log.Errorw("failed to reserve spot", "error", err)
						select {
						case <-ctx.Done():
							return
						case <-time.After(waitTime.Duration()):
							continue
						}
					}

					log.Infow("reserved spot", "reservation", reservation)
					waitTime.Reset()
					select {
					case <-ctx.Done():
						return
					case <-time.After(waitTime.Min):
						continue
					}
				}
			}
		}()
	}

	<-ctx.Done()
	return nil
}
