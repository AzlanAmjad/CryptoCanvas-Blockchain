package core

import "io"

type Encoder[T any] interface {
	Encode(w io.Writer, v T) error
}

type Decoder[T any] interface {
	Decode(r io.Reader, v T) error
}

type BlockEncoder struct{}

func NewBlockEncoder() *BlockEncoder {
	return &BlockEncoder{}
}

func (e *BlockEncoder) Encode(w io.Writer, v *Block) error {
	return nil
}

type BlockDecoder struct{}

func NewBlockDecoder() *BlockDecoder {
	return &BlockDecoder{}
}

func (d *BlockDecoder) Decode(r io.Reader, v *Block) error {
	return nil
}
