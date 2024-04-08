package core

import (
	"testing"

	crypto "github.com/AzlanAmjad/DreamscapeCanvas-Blockchain/cryptography"
	"github.com/stretchr/testify/assert"
)

func TestAccountStateTransferFail(t *testing.T) {
	// create a new account state
	state := NewAccountStates()

	// Try and transfer from arbitrary accounts to another
	fromPrivKey := crypto.GeneratePrivateKey()
	fromPubKey := fromPrivKey.GetPublicKey()
	fromAddress := fromPubKey.GetAddress()

	toPrivKey := crypto.GeneratePrivateKey()
	toPubKey := toPrivKey.GetPublicKey()
	toAddress := toPubKey.GetAddress()

	err := state.Transfer(fromAddress, toAddress, 100)
	assert.NotNil(t, err)
}

func TestSAccountStateTransferPass(t *testing.T) {
	// create a new state storage
	state := NewAccountStates()

	// create sender account and add balance to that account
	fromPrivKey := crypto.GeneratePrivateKey()
	fromPubKey := fromPrivKey.GetPublicKey()
	fromAddress := fromPubKey.GetAddress()

	state.AddBalance(fromAddress, 300)

	// transfer some of the funds
	toPrivKey := crypto.GeneratePrivateKey()
	toPubKey := toPrivKey.GetPublicKey()
	toAddress := toPubKey.GetAddress()

	// transfer the funds
	err := state.Transfer(fromAddress, toAddress, 200)
	assert.Nil(t, err)
}
