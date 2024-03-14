package core

import (
	"bytes"
	"encoding/binary"
	"fmt"

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
	fmt.Println("Creating a new LevelDBStorage at path:", dbPath)
	db, err := leveldb.OpenFile(dbPath, nil)
	if err != nil {
		return nil, err
	}
	return &LevelDBStorage{Init: true, DB: db}, nil
}

// Put adds a block to the database.
func (s *LevelDBStorage) Put(block *Block, enc Encoder[*Block]) error {
	// NOTE: TO BE IMPLEMENTED, MUST HAVE BLOCK ENCODING IMPLEMENTED FIRST
	key := make([]byte, 4)
	binary.LittleEndian.PutUint32(key, block.Header.Index)
	value := bytes.Buffer{}
	block.Encode(&value, NewBlockEncoder())
	err := s.DB.Put(key, value.Bytes(), nil)
	if err != nil {
		return err
	}
	return nil
}

// Get returns a block from the database.
func (s *LevelDBStorage) Get(index uint32, dec Decoder[*Block]) (*Block, error) {
	// NOTE: TO BE IMPLEMENTED, MUST HAVE BLOCK DECODING IMPLEMENTED FIRST
	key := make([]byte, 4)
	binary.LittleEndian.PutUint32(key, index)
	value, err := s.DB.Get(key, nil)
	if err != nil {
		return nil, err
	}
	block := NewBlock()
	block.Decode(bytes.NewReader(value), NewBlockDecoder())
	return block, nil
}

// Shutdown closes the database.
func (s *LevelDBStorage) Shutdown() {
	s.DB.Close()
}
