package core

import (
	"encoding/hex"
	"fmt"
	"sync"

	types "github.com/AzlanAmjad/DreamscapeCanvas-Blockchain/data-types"
)

type Account struct {
	Address types.Address
	Balance types.CurrencyAmount
}

type AccountStates struct {
	mu sync.RWMutex
	// Account state is a map of account addresses to account currency holding value.
	state map[types.Address]*Account
}

func NewAccountStates() *AccountStates {
	return &AccountStates{
		state: make(map[types.Address]*Account),
	}
}

func (as *AccountStates) GetAccount(address types.Address) (*Account, error) {
	as.mu.RLock()
	defer as.mu.RUnlock()

	if _, ok := as.state[address]; !ok {
		return nil, fmt.Errorf("address %s does not exist in the account state", hex.EncodeToString(address[:]))
	}

	return as.state[address], nil
}

func (as *AccountStates) GetBalance(address types.Address) (types.CurrencyAmount, error) {
	as.mu.RLock()
	defer as.mu.RUnlock()

	if _, ok := as.state[address]; !ok {
		return 0, fmt.Errorf("address %s does not exist in the account state", hex.EncodeToString(address[:]))
	}

	return as.state[address].Balance, nil
}

func (as *AccountStates) AddBalance(address types.Address, amount types.CurrencyAmount) {
	as.mu.Lock()
	defer as.mu.Unlock()

	// Check if the address exists in the map
	if _, ok := as.state[address]; !ok {
		as.state[address] = &Account{Address: address, Balance: 0} // Initialize the account with 0 balance
	}
	as.state[address].Balance += amount
}

func (as *AccountStates) SubtractBalance(address types.Address, amount types.CurrencyAmount) error {
	as.mu.Lock()
	defer as.mu.Unlock()

	// Check if the account exists in the map
	if _, ok := as.state[address]; !ok {
		return fmt.Errorf("address %s does not exist in the account state", hex.EncodeToString(address[:]))
	}

	// Check if the account has enough balance
	if as.state[address].Balance < amount {
		return fmt.Errorf("insufficient balance in account %s", hex.EncodeToString(address[:]))
	}

	as.state[address].Balance -= amount

	return nil
}

func (as *AccountStates) Transfer(from types.Address, to types.Address, amount types.CurrencyAmount) error {
	// Try and subtract from the sender's account
	if err := as.SubtractBalance(from, amount); err != nil {
		return err
	}

	// Add to the receiver's account, if it doesn't exist, it is created
	as.AddBalance(to, amount)

	return nil
}
