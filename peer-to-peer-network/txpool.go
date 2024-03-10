package network

import (
	"container/heap"

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
	priorityQueue     PriorityQueue
}

func NewTxPool() *TxPool {
	pq := make(PriorityQueue, 0)
	heap.Init(&pq)
	return &TxPool{
		transactions:      make(map[types.Hash]*core.Transaction),
		TransactionHasher: core.NewTransactionHasher(), // Create a new default transaction hasher.
		priorityQueue:     pq,                          // Initialize the priority queue. Used to order transactions in the mempool based on when they were first added.
	}
}

func (c *TxPool) Len() int {
	return len(c.transactions)
}

func (c *TxPool) Flush() {
	// reset the mempool map
	c.transactions = make(map[types.Hash]*core.Transaction)
	// reset the priority queue
	c.priorityQueue = make(PriorityQueue, 0)
	heap.Init(&c.priorityQueue)
}

// Adds a transaction to the mempool. Caller is responsible for checking the validity of the transaction.
// Caller is also responsible for checking if the transaction is already in the mempool.
func (c *TxPool) Add(tx *core.Transaction) error {
	c.transactions[tx.GetHash(c.TransactionHasher)] = tx // add to map
	heap.Push(&c.priorityQueue, tx)                      // add to priority queue
	return nil
}

func (c *TxPool) Get(hash types.Hash) *core.Transaction {
	return c.transactions[hash]
}

func (c *TxPool) Has(hash types.Hash) bool {
	_, ok := c.transactions[hash]
	return ok
}

// function to read all transactions from the mempool
// highest priority transactions are read first and are at the top of the list
func (c *TxPool) GetTransactions() []*core.Transaction {
	transactions := make([]*core.Transaction, 0, len(c.transactions))
	for c.priorityQueue.Len() > 0 {
		tx := heap.Pop(&c.priorityQueue).(*core.Transaction)
		transactions = append(transactions, tx)
	}
	return transactions
}
