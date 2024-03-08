package core

import "io"

type Encoder[T any] interface {
	Encode(w io.Writer, v T) error
}

type Decoder[T any] interface {
	Decode(r io.Reader, v T) error
}

type BlockEncoder struct{}

type BlockDecoder struct{}
