package core

import (
	"github.com/syndtr/goleveldb/leveldb"
)

type Storage interface {
	Put(block *Block, enc Encoder[*Block]) error
	Get(index uint32, dec Decoder[*Block]) (*Block, error)
	Shutdown()
}

// LevelDBStorage is a storage implementation using LevelDB.
type LevelDBStorage struct {
	Init bool
	DB   *leveldb.DB
}

// NewLevelDBStorage creates a new LevelDBStorage.
func NewLevelDBStorage(dbPath string) (*LevelDBStorage, error) {
	db, err := leveldb.OpenFile(dbPath, nil)
	if err != nil {
		return nil, err
	}
	return &LevelDBStorage{Init: true, DB: db}, nil
}

// Put adds a block to the database.
func (s *LevelDBStorage) Put(block *Block, enc Encoder[*Block]) error {
	// NOTE: TO BE IMPLEMENTED, MUST HAVE BLOCK ENCODING IMPLEMENTED FIRST
	return nil
}

// Get returns a block from the database.
func (s *LevelDBStorage) Get(index uint32, dec Decoder[*Block]) (*Block, error) {
	// NOTE: TO BE IMPLEMENTED, MUST HAVE BLOCK DECODING IMPLEMENTED FIRST
	return nil, nil
}

// Shutdown closes the database.
func (s *LevelDBStorage) Shutdown() {
	s.DB.Close()
}
