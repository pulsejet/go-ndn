package encoding

import (
	"errors"
	"fmt"
	"io"
)

// Buffer is a buffer of bytes
type Buffer []byte

// Wire is a collection of Buffer. May be allocated in non-contiguous memory.
type Wire []Buffer

func (w Wire) Join() []byte {
	if len(w) == 0 {
		return []byte{}
	} else if len(w) == 1 {
		return w[0]
	}

	n := 0
	for _, v := range w {
		n += len(v)
	}

	b := make([]byte, n)
	bp := copy(b, w[0])
	for _, v := range w[1:] {
		bp += copy(b[bp:], v)
	}
	return b
}

type ErrFormat struct {
	Msg string
}

func (e ErrFormat) Error() string {
	return e.Msg
}

type ErrNotFound struct {
	Key string
}

func (e ErrNotFound) Error() string {
	return e.Key + ": not found"
}

// ParseReader is an interface operating on Buffer and Wire
type ParseReader interface {
	io.Reader
	io.ByteScanner

	// ReadWire reads a list of buffers in place without copy.
	// It always tries to read the required length of bytes.
	ReadWire(l int) (Wire, error)

	// Range returns a wire that contains the bytes between start and end, without copy.
	Range(start, end int) Wire

	// Pos returns the current position in the buffer/wire.
	Pos() int

	// Length returns the length of the buffer/wire.
	Length() int

	// Skip skips the next n bytes.
	Skip(n int) error

	// Delegate returns a new ParseReader that starts from the current position with length l.
	// The result is equivalent to the following:
	//
	//   start := r.Pos()
	//   r.Skip(l)
	//   return r.Range(start, r.Pos())
	Delegate(l int) ParseReader
}

type ErrUnrecognizedField struct {
	TypeNum TLNum
}

func (e ErrUnrecognizedField) Error() string {
	return fmt.Sprintf("There exists an unrecognized field that has a critical type number: %d", e.TypeNum)
}

var ErrBufferOverflow = errors.New("Buffer overflow when parsing. One of the TLV Length is wrong")

var ErrIncorrectDigest = errors.New("The sha256 digest is missing or incorrect")

type ErrSkipRequired struct {
	TypeNum TLNum
}

func (e ErrSkipRequired) Error() string {
	return fmt.Sprintf("The required field of type %d is missing in the wire", e.TypeNum)
}

type ErrFailToParse struct {
	TypeNum TLNum
	Err     error
}

func (e ErrFailToParse) Error() string {
	return fmt.Sprintf("Failed to parse field %d: %v", e.TypeNum, e.Err)
}

func (e ErrFailToParse) Unwrap() error {
	return e.Err
}

type ErrUnexpected struct {
	Err error
}

func (e ErrUnexpected) Error() string {
	return fmt.Sprintf("Unexpected error happened in parsing: %v", e.Err)
}

func (e ErrUnexpected) Unwrap() error {
	return e.Err
}

// PlaceHolder is an empty structure that used to give names of procedure arguments.
type PlaceHolder struct{}
