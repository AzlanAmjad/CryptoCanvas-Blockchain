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
	GetBlocks(start, end uint32, dec Decoder[*Block]) ([]*Block, error)
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
	key := make([]byte, 4)
	binary.LittleEndian.PutUint32(key, block.Header.Index)
	value := bytes.Buffer{}
	err := block.Encode(&value, enc)
	if err != nil {
		return err
	}
	err = s.DB.Put(key, value.Bytes(), nil)
	if err != nil {
		return err
	}
	return nil
}

// Get returns a block from the database.
func (s *LevelDBStorage) Get(index uint32, dec Decoder[*Block]) (*Block, error) {
	key := make([]byte, 4)
	binary.LittleEndian.PutUint32(key, index)
	value, err := s.DB.Get(key, nil)
	if err != nil {
		return nil, err
	}
	block := NewBlock()
	err = block.Decode(bytes.NewReader(value), dec)
	if err != nil {
		return nil, err
	}
	return block, nil
}

// GetBlocks returns all blocks from start to end
// end is not inclusive
func (s *LevelDBStorage) GetBlocks(start, end uint32, dec Decoder[*Block]) ([]*Block, error) {
	var blocks []*Block
	for i := start; i < end; i++ {
		block, err := s.Get(i, dec)
		if err != nil {
			return nil, err
		}
		blocks = append(blocks, block)
	}
	return blocks, nil
}

// Shutdown closes the database.
func (s *LevelDBStorage) Shutdown() {
	s.DB.Close()
}
