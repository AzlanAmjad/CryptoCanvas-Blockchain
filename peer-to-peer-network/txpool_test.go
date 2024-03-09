package network

import (
	"testing"

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
