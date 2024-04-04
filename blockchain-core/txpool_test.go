package core

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTxPool(t *testing.T) {
	txPool := NewTxPool(100)

	assert.Equal(t, 0, txPool.PendingLen())

	tx1 := NewTransaction([]byte("tx1"))
	tx2 := NewTransaction([]byte("tx2"))

	txPool.Add(tx1)
	assert.Equal(t, 1, txPool.PendingLen())
	assert.True(t, txPool.PendingHas(tx1.GetHash(txPool.Pending.TransactionHasher)))
	assert.False(t, txPool.PendingHas(tx2.GetHash(txPool.Pending.TransactionHasher)))

	txPool.Add(tx2)
	assert.Equal(t, 2, txPool.PendingLen())
	assert.True(t, txPool.PendingHas(tx2.GetHash(txPool.Pending.TransactionHasher)))

	txPool.FlushPending()
	assert.Equal(t, 0, txPool.PendingLen())
	assert.False(t, txPool.PendingHas(tx1.GetHash(txPool.Pending.TransactionHasher)))
	assert.False(t, txPool.PendingHas(tx2.GetHash(txPool.Pending.TransactionHasher)))
}

func TestGetTransactionInSortedOrder(t *testing.T) {
	txPool := NewTxPool(100)

	tx1 := NewTransaction([]byte("tx1"))
	tx1.SetFirstSeen(time.Now().UnixNano()) // Set first seen time to current time;

	time.Sleep(1 * time.Millisecond)

	tx2 := NewTransaction([]byte("tx2"))
	tx2.SetFirstSeen(time.Now().UnixNano())

	time.Sleep(1 * time.Millisecond)

	tx3 := NewTransaction([]byte("tx3"))
	tx3.SetFirstSeen(time.Now().UnixNano())

	// Add tx in random order
	txPool.Add(tx2)
	txPool.Add(tx3)
	txPool.Add(tx1)

	transactions := txPool.GetPendingTransactions()

	// Check if transactions are in sorted order
	assert.Equal(t, tx1, transactions[0])
	assert.Equal(t, tx2, transactions[1])
	assert.Equal(t, tx3, transactions[2])
}

func TestTxPoolPruning(t *testing.T) {
	// pool with maxLen 2
	txPool := NewTxPool(2)

	// add 3 transactions
	tx1 := NewTransaction([]byte("tx1"))
	tx2 := NewTransaction([]byte("tx2"))
	tx3 := NewTransaction([]byte("tx3"))

	txPool.Add(tx1)
	txPool.Add(tx2)
	txPool.Add(tx3)

	// tx1 should be pruned
	assert.False(t, txPool.AllHas(tx1.GetHash(txPool.All.TransactionHasher)))
	assert.Equal(t, 2, txPool.AllLen())
}
