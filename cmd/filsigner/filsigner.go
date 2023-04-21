package main

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	client2 "github.com/data-preservation-programs/filsigner-relayed/client"
	"github.com/data-preservation-programs/filsigner-relayed/config"
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
	"net/http"
	"os"
)

func healthHandler(w http.ResponseWriter, r *http.Request) {
	// Set response header to 200 OK
	w.WriteHeader(http.StatusOK)
	// Write a message to the response body
	fmt.Fprint(w, "OK")
}

func main() {
	log := logging.Logger("server")
	address.CurrentNetwork = address.Mainnet
	allowedRequestersArg := new(cli.StringSlice)
	signKeysArg := new(cli.StringSlice)
	identityKeyArg := new(string)
	relayInfos := new(cli.StringSlice)

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
					&cli.StringSliceFlag{
						Name:        "relay-info",
						Usage:       "The relay info to use to connect to the allowed requesters - this will override the default relay servers from SPADE",
						Destination: relayInfos,
						EnvVars:     []string{"RELAY_INFOS"},
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

					var relays []peer.AddrInfo
					if len(relayInfos.Value()) == 0 {
						relays = config.GetDefaultRelayInfo()
					} else {
						relays = make([]peer.AddrInfo, len(relayInfos.Value()))
						for i, relayInfo := range relayInfos.Value() {
							relay, err := peer.AddrInfoFromString(relayInfo)
							if err != nil {
								return errors.Wrapf(err, "cannot decode relay info %s", relayInfo)
							}

							relays[i] = *relay
						}
					}

					client, err := client2.NewClient(identityKey, relays)
					if err != nil {
						return errors.Wrap(err, "cannot create client")
					}

					proposal := filmarket.DealProposal{
						PieceCID:             cid.MustParse("baga6ea4seaqgvktrw7sh3ypsuai76csagofcgnq6xlyulk5wjcunqsx6pg7dqfa"),
						PieceSize:            256,
						VerifiedDeal:         true,
						Client:               clientAddr,
						Provider:             address.TestAddress,
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
					&cli.StringSliceFlag{
						Name:        "relay-info",
						Usage:       "[Local testing only] The relay info to use to connect to the allowed requesters - this will override the default relay servers from SPADE",
						Destination: relayInfos,
						EnvVars:     []string{"RELAY_INFOS"},
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

					var relays []peer.AddrInfo
					if len(relayInfos.Value()) == 0 {
						relays = config.GetDefaultRelayInfo()
					} else {
						relays = make([]peer.AddrInfo, len(relayInfos.Value()))
						for i, relayInfo := range relayInfos.Value() {
							relay, err := peer.AddrInfoFromString(relayInfo)
							if err != nil {
								return errors.Wrapf(err, "cannot decode relay info %s", relayInfo)
							}

							relays[i] = *relay
						}
					}

					server, err := server.NewServer(identityKey, allowedRequesters, signKeysArg.Value(), relays)
					if err != nil {
						return errors.Wrap(err, "cannot create new server")
					}

					go func() {
						// Register the healthHandler function for the /health route
						http.HandleFunc("/healthz", healthHandler)

						// Start the HTTP server on port 8088
						fmt.Println("Listening on :8088...")
						err := http.ListenAndServe(":8088", nil)
						if err != nil {
							log.Fatal(err)
						}
					}()

					err = server.Start(c.Context)
					if err != nil {
						return errors.Wrap(err, "cannot start server")
					}

					return nil
				},
			},
			{
				Name:  "generate-peer",
				Usage: "generate a new peer id with private key",
				Action: func(c *cli.Context) error {
					privateStr, publicStr, peerStr, err := GenerateNewPeer()
					if err != nil {
						return errors.Wrap(err, "cannot generate new peer")
					}

					//nolint:forbidigo
					{
						fmt.Println("New peer generated using ed25519, keys are encoded in base64")
						fmt.Println("peer id:     ", peerStr.String())
						fmt.Println("public key:  ", publicStr)
						fmt.Println("private key: ", privateStr)
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
func GenerateNewPeer() (string, string, peer.ID, error) {
	private, public, err := crypto.GenerateEd25519Key(rand.Reader)
	if err != nil {
		return "", "", "", errors.Wrap(err, "cannot generate new peer")
	}

	peerID, err := peer.IDFromPublicKey(public)
	if err != nil {
		return "", "", "", errors.Wrap(err, "cannot generate peer id")
	}

	privateBytes, err := crypto.MarshalPrivateKey(private)
	if err != nil {
		return "", "", "", errors.Wrap(err, "cannot marshal private key")
	}

	privateStr := base64.StdEncoding.EncodeToString(privateBytes)

	publicBytes, err := crypto.MarshalPublicKey(public)
	if err != nil {
		return "", "", "", errors.Wrap(err, "cannot marshal public key")
	}

	publicStr := base64.StdEncoding.EncodeToString(publicBytes)
	return privateStr, publicStr, peerID, nil
}
