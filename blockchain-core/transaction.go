package core

import (
	"encoding/binary"
	"io"
)

type Transaction struct {
	Data []byte
}

func (tx *Transaction) BinaryEncode(w io.Writer) error {
	err := binary.Write(w, binary.LittleEndian, &tx.Data)
	if err != nil {
		return err
	}
	return nil
}

func (tx *Transaction) BinaryDecode(r io.Reader) error {
	err := binary.Read(r, binary.LittleEndian, &tx.Data)
	if err != nil {
		return err
	}
	return nil
}
