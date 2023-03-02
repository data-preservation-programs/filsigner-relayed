package main

import (
	"encoding/base64"
	client2 "github.com/data-preservation-programs/filsigner-relayed/client"
	"github.com/data-preservation-programs/filsigner-relayed/server"
	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-state-types/abi"
	filmarket "github.com/filecoin-project/go-state-types/builtin/v9/market"
	"github.com/ipfs/go-cid"
	logging "github.com/ipfs/go-log/v2"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"
	"os"
)

func main() {
	log := logging.Logger("server")
	allowedRequestersArg := new(cli.StringSlice)
	signKeysArg := new(cli.StringSlice)
	identityKeyArg := new(string)

	destination := new(string)
	client := new(string)

	app := &cli.App{
		Name: "filsigner",
		Commands: []*cli.Command{
			{
				Name:  "test",
				Usage: "Request a signature from the filsigner server for a test proposal",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:        "identity-key",
						Aliases:     []string{"k"},
						Usage:       "The base64 encoded private key of the peer to use as the identity",
						Destination: identityKeyArg,
						EnvVars:     []string{"IDENTITY_KEY"},
						Required:    true,
					},
					&cli.StringFlag{
						Name:        "destination",
						Aliases:     []string{"d"},
						Usage:       "The peer ID to send the deal proposal to",
						Destination: destination,
						Required:    true,
					},
					&cli.StringFlag{
						Name:        "client",
						Aliases:     []string{"c"},
						Usage:       "The client address to use in the deal proposal",
						Destination: client,
						Required:    true,
					},
				},
				Action: func(c *cli.Context) error {
					identityKeyBytes, err := base64.StdEncoding.DecodeString(*identityKeyArg)
					if err != nil {
						return errors.Wrap(err, "cannot decode identity key")
					}

					identityKey, err := crypto.UnmarshalPrivateKey(identityKeyBytes)
					if err != nil {
						return errors.Wrap(err, "cannot unmarshal identity private key")
					}

					destinationPeer, err := peer.Decode(*destination)
					if err != nil {
						return errors.Wrap(err, "cannot decode destination peer ID")
					}

					clientAddr, err := address.NewFromString(*client)
					if err != nil {
						return errors.Wrap(err, "cannot decode client address")
					}

					client, err := client2.NewClient(identityKey)
					if err != nil {
						return errors.Wrap(err, "cannot create client")
					}

					proposal := filmarket.DealProposal{
						PieceCID:             cid.MustParse("baga6ea4seaqgvktrw7sh3ypsuai76csagofcgnq6xlyulk5wjcunqsx6pg7dqfa"),
						PieceSize:            256,
						VerifiedDeal:         true,
						Client:               clientAddr,
						Provider:             address.Undef,
						Label:                filmarket.EmptyDealLabel,
						StartEpoch:           0,
						EndEpoch:             0,
						StoragePricePerEpoch: abi.TokenAmount{},
						ProviderCollateral:   abi.TokenAmount{},
						ClientCollateral:     abi.TokenAmount{},
					}
					
					signature, err := client.SignProposal(c.Context, destinationPeer, proposal)
					if err != nil {
						return errors.Wrap(err, "cannot sign proposal")
					}

					signatureBytes, err := signature.MarshalBinary()
					if err != nil {
						return errors.Wrap(err, "cannot marshal signature")
					}

					log.Infof("Signature(base64): %s", base64.StdEncoding.EncodeToString(signatureBytes))
					return nil
				},
			},
			{
				Name:  "run",
				Usage: "Run the filsigner server to sign deal proposals",
				Flags: []cli.Flag{
					&cli.StringSliceFlag{
						Name:        "allowed-requester",
						Aliases:     []string{"r"},
						Usage:       "The peer ID of the allowed requester",
						Destination: allowedRequestersArg,
						EnvVars:     []string{"ALLOWED_REQUESTERS"},
						Required:    true,
					},
					&cli.StringFlag{
						Name:        "identity-key",
						Aliases:     []string{"k"},
						Usage:       "The base64 encoded private key of the peer to use as the identity",
						Destination: identityKeyArg,
						EnvVars:     []string{"IDENTITY_KEY"},
						Required:    true,
					},
					&cli.StringSliceFlag{
						Name:        "sign-key",
						Aliases:     []string{"s"},
						Usage:       "The private key of the address to sign with",
						Destination: signKeysArg,
						EnvVars:     []string{"SIGN_KEYS"},
						Required:    true,
					},
				},
				Action: func(c *cli.Context) error {
					identityKeyBytes, err := base64.StdEncoding.DecodeString(*identityKeyArg)
					if err != nil {
						return errors.Wrap(err, "cannot decode identity key")
					}

					identityKey, err := crypto.UnmarshalPrivateKey(identityKeyBytes)
					if err != nil {
						return errors.Wrap(err, "cannot unmarshal identity private key")
					}

					allowedRequesters := make([]peer.ID, len(allowedRequestersArg.Value()))
					for i, allowedRequester := range allowedRequestersArg.Value() {
						allowedRequesters[i], err = peer.Decode(allowedRequester)
						if err != nil {
							return errors.Wrapf(err, "cannot decode allowed requester %s", allowedRequester)
						}
					}

					server, err := server.NewServer(identityKey, allowedRequesters, signKeysArg.Value())
					if err != nil {
						return errors.Wrap(err, "cannot create new server")
					}

					err = server.Start(c.Context)
					if err != nil {
						return errors.Wrap(err, "cannot start server")
					}

					return nil
				},
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatalf("Failed to run filsigner: %v", err)
	}
}
