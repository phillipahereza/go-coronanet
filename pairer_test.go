// go-coronanet - Coronavirus social distancing network
// Copyright (c) 2020 Péter Szilágyi. All rights reserved.

package coronanet

import (
	"bytes"
	"context"
	"testing"

	"github.com/coronanet/go-coronanet/tornet"
)

// Tests that basic pairing works.
func TestPairing(t *testing.T) {
	// Create two identities, one for initiating pairing and one for joining
	initKeyRing, _ := tornet.GenerateKeyRing()
	joinKeyRing, _ := tornet.GenerateKeyRing()

	initRemote := tornet.RemoteKeyRing{
		Identity: initKeyRing.Identity.Public(),
		Address:  initKeyRing.Addresses[0].Public(),
	}
	joinRemote := tornet.RemoteKeyRing{
		Identity: joinKeyRing.Identity.Public(),
		Address:  joinKeyRing.Addresses[0].Public(),
	}
	// Initiate a pairing session and join it with the other identity
	gateway := tornet.NewMockGateway()

	initPairer, secret, address, err := newPairingServer(gateway, initRemote)
	if err != nil {
		t.Fatalf("failed to initiate pairing: %v", err)
	}
	joinPairer, err := newPairingClient(gateway, joinRemote, secret, address)
	if err != nil {
		t.Fatalf("failed to join pairing: %v", err)
	}
	// Wait for both to finish
	joinPub, err := initPairer.wait(context.TODO())
	if err != nil {
		t.Fatalf("server side pairing failed: %v", err)
	}
	initPub, err := joinPairer.wait(context.TODO())
	if err != nil {
		t.Fatalf("client side pairing failed: %v", err)
	}
	// Ensure the exchanged secrets match
	if !bytes.Equal(initPub.Identity, initRemote.Identity) {
		t.Errorf("initer identity mismatch: have %x, want %x", initPub.Identity, initRemote.Identity)
	}
	if !bytes.Equal(initPub.Address, initRemote.Address) {
		t.Errorf("initer address mismatch: have %x, want %x", initPub.Address, initRemote.Address)
	}
	if !bytes.Equal(joinPub.Identity, joinRemote.Identity) {
		t.Errorf("joiner identity mismatch: have %x, want %x", joinPub.Identity, joinRemote.Identity)
	}
	if !bytes.Equal(joinPub.Address, joinRemote.Address) {
		t.Errorf("joiner address mismatch: have %x, want %x", joinPub.Address, joinRemote.Address)
	}
}
