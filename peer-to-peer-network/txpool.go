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
	All       *SortedTxMap // all transactions we have ever seen
	Pending   *SortedTxMap // pending transactions that we have yet to consume
	MaxLength int          // prune the oldest transactions (smallest timestamps) in All map when it gets too big
}

func NewTxPool(maxLength int) *TxPool {
	return &TxPool{
		All:       NewSortedTxMap(),
		Pending:   NewSortedTxMap(),
		MaxLength: maxLength,
	}
}

func (t *TxPool) Add(tx *core.Transaction) error {
	// add to pending
	err := t.Pending.Add(tx)
	if err != nil {
		return err
	}
	// add to all
	err = t.All.Add(tx)
	if err != nil {
		return err
	}
	// check if all is longer than max length
	if t.All.Len() > t.MaxLength {
		// remove the oldest transaction
		t.All.RemoveOldestTransaction()
	}
	return nil
}

func (t *TxPool) PendingLen() int {
	return t.Pending.Len()
}

func (t *TxPool) AllLen() int {
	return t.All.Len()
}

func (t *TxPool) FlushPending() {
	t.Pending.Flush()
}

func (t *TxPool) GetPendingTransactions() []*core.Transaction {
	return t.Pending.GetTransactions()
}

func (t *TxPool) GetAllTransactions() []*core.Transaction {
	return t.All.GetTransactions()
}

func (t *TxPool) PendingHas(hash types.Hash) bool {
	return t.Pending.Has(hash)
}

func (t *TxPool) AllHas(hash types.Hash) bool {
	return t.All.Has(hash)
}

// data structures that contain transactions in a sorted manner
// accessible by hash like a map
// basically a priority queue layered with a hash map
type SortedTxMap struct {
	transactions      map[types.Hash]*core.Transaction
	TransactionHasher core.Hasher[*core.Transaction]
	priorityQueue     PriorityQueue
}

// new sorted tx map
func NewSortedTxMap() *SortedTxMap {
	pq := make(PriorityQueue, 0)
	heap.Init(&pq)
	return &SortedTxMap{
		transactions:      make(map[types.Hash]*core.Transaction),
		TransactionHasher: core.NewTransactionHasher(), // Create a new default transaction hasher.
		priorityQueue:     pq,                          // Initialize the priority queue. Used to order transactions in the mempool based on when they were first added.
	}
}

// length of transactions
func (c *SortedTxMap) Len() int {
	return len(c.transactions)
}

// flush the mempool
func (c *SortedTxMap) Flush() {
	// reset the mempool map
	c.transactions = make(map[types.Hash]*core.Transaction)
	// reset the priority queue
	c.priorityQueue = make(PriorityQueue, 0)
	heap.Init(&c.priorityQueue)
}

// Adds a transaction to the mempool. Caller is responsible for checking the validity of the transaction.
// Caller is also responsible for checking if the transaction is already in the mempool.
func (c *SortedTxMap) Add(tx *core.Transaction) error {
	c.transactions[tx.GetHash(c.TransactionHasher)] = tx // add to map
	heap.Push(&c.priorityQueue, tx)                      // add to priority queue
	return nil
}

// get a transaction from the mempool, using the hash of the transaction
func (c *SortedTxMap) Get(hash types.Hash) *core.Transaction {
	return c.transactions[hash]
}

// checks if the transaction already exists in the mempool
func (c *SortedTxMap) Has(hash types.Hash) bool {
	_, ok := c.transactions[hash]
	return ok
}

// function to read all transactions from the mempool
// highest priority transactions are read first and are at the top of the list
func (c *SortedTxMap) GetTransactions() []*core.Transaction {
	transactions := make([]*core.Transaction, 0, len(c.transactions))
	for c.priorityQueue.Len() > 0 {
		tx := heap.Pop(&c.priorityQueue).(*core.Transaction)
		transactions = append(transactions, tx)
	}
	return transactions
}

// removes the oldest transaction (smallest timestamp)
func (c *SortedTxMap) RemoveOldestTransaction() {
	// get the oldest transaction
	tx := heap.Pop(&c.priorityQueue).(*core.Transaction)
	// remove it from the map
	delete(c.transactions, tx.GetHash(c.TransactionHasher))
}
