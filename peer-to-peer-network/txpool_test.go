package network

import (
	"testing"
	"time"

	core "github.com/AzlanAmjad/DreamscapeCanvas-Blockchain/blockchain-core"
	"github.com/stretchr/testify/assert"
)

func TestTxPool(t *testing.T) {
	txPool := NewTxPool()

	assert.Equal(t, 0, txPool.Len())

	tx1 := core.NewTransaction([]byte("tx1"))
	tx2 := core.NewTransaction([]byte("tx2"))

	txPool.Add(tx1)
	assert.Equal(t, 1, txPool.Len())
	assert.True(t, txPool.Has(tx1.GetHash(txPool.TransactionHasher)))
	assert.False(t, txPool.Has(tx2.GetHash(txPool.TransactionHasher)))

	txPool.Add(tx2)
	assert.Equal(t, 2, txPool.Len())
	assert.True(t, txPool.Has(tx2.GetHash(txPool.TransactionHasher)))

	txPool.Flush()
	assert.Equal(t, 0, txPool.Len())
	assert.False(t, txPool.Has(tx1.GetHash(txPool.TransactionHasher)))
	assert.False(t, txPool.Has(tx2.GetHash(txPool.TransactionHasher)))
}

func TestGetTransactionInSortedOrder(t *testing.T) {
	txPool := NewTxPool()

	tx1 := core.NewTransaction([]byte("tx1"))
	tx1.SetFirstSeen(time.Now().UnixNano()) // Set first seen time to current time;

	time.Sleep(1 * time.Millisecond)

	tx2 := core.NewTransaction([]byte("tx2"))
	tx2.SetFirstSeen(time.Now().UnixNano())

	time.Sleep(1 * time.Millisecond)

	tx3 := core.NewTransaction([]byte("tx3"))
	tx3.SetFirstSeen(time.Now().UnixNano())

	// Add tx in random order
	txPool.Add(tx2)
	txPool.Add(tx3)
	txPool.Add(tx1)

	transactions := txPool.GetTransactions()

	// Check if transactions are in sorted order
	assert.Equal(t, tx1, transactions[0])
	assert.Equal(t, tx2, transactions[1])
	assert.Equal(t, tx3, transactions[2])
}
