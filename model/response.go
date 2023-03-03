package model

type StatusCode uint64

const (
	Success StatusCode = iota
	UnauthorizedRequester
	ReadStreamError
	DecodeRequestError
	EncodeRequestError
	ProposalRemarshalMismatch
	WalletKeyNotFound
	WalletSignError
	MarshalSignatureError
	EncodeResponseError
)

var StatusCodeString = []string{
	"Success",
	"UnauthorizedRequester",
	"ReadStreamError",
	"DecodeRequestError",
	"EncodeRequestError",
	"ProposalRemarshalMismatch",
	"WalletKeyNotFound",
	"WalletSignError",
	"MarshalSignatureError",
	"EncodeResponseError",
}

//go:generate go run github.com/hannahhoward/cbor-gen-for --map-encoding SignerResponse

type SignerResponse struct {
	Code      StatusCode
	Message   string
	Signature []byte
}
