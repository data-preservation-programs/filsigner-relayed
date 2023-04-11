package server

import (
	"github.com/filecoin-project/go-address"
	"testing"
)

func TestResolveShortID(t *testing.T) {
	address.CurrentNetwork = address.Mainnet
	addr, err := address.NewFromString("f2kmbjvz7vagl2z6pfrbjoggrkjofxspp7cqtw2zy")
	if err != nil {
		t.Fatalf("err is not null: %v", err)
	}

	shortAddr, err := resolveShortID(addr)
	if err != nil {
		t.Fatalf("err is not null: %v", err)
	}

	if shortAddr.String() != "f0123261" {
		t.Fatalf("short addresss is incorrect: %v", shortAddr.String())
	}
}
