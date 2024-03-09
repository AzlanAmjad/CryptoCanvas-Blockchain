package network

import (
	core "github.com/AzlanAmjad/DreamscapeCanvas-Blockchain/blockchain-core"
	types "github.com/AzlanAmjad/DreamscapeCanvas-Blockchain/data-types"
)

// Transaction Mempool
// Maintain list of transactions that are not yet included in a block.
// If a client makes a transaction / posts a transaction to the node, node will check validity of the transaction
// and add it to the mempool, and broadcast it to the network, each node will add it to their mempool and
// do the same. When a new block is created, the transactions in the mempool will be included in the block.

type TxPool struct {
	transactions      map[types.Hash]*core.Transaction
	TransactionHasher core.Hasher[*core.Transaction]
}

func NewTxPool() *TxPool {
	return &TxPool{
		transactions:      make(map[types.Hash]*core.Transaction),
		TransactionHasher: core.NewTransactionHasher(), // Create a new default transaction hasher.
	}
}

func (c *TxPool) Len() int {
	return len(c.transactions)
}

func (c *TxPool) Flush() {
	c.transactions = make(map[types.Hash]*core.Transaction)
}

// Adds a transaction to the mempool. Caller is responsible for checking the validity of the transaction.
// Caller is also responsible for checking if the transaction is already in the mempool.
func (c *TxPool) Add(tx *core.Transaction) error {
	c.transactions[tx.GetHash(c.TransactionHasher)] = tx
	return nil
}

func (c *TxPool) Get(hash types.Hash) *core.Transaction {
	return c.transactions[hash]
}

func (c *TxPool) Has(hash types.Hash) bool {
	_, ok := c.transactions[hash]
	return ok
}
