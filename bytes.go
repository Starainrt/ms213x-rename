package main

import (
	"fmt"
	"io"
)

type ByteSlice struct {
	data []byte
	pos  int
}

func NewByteSlice(b []byte) *ByteSlice {
	return &ByteSlice{data: b}
}

func (bs *ByteSlice) Read(p []byte) (int, error) {
	if bs.pos >= len(bs.data) {
		return 0, io.EOF
	}
	n := copy(p, bs.data[bs.pos:])
	bs.pos += n
	return n, nil
}

func (bs *ByteSlice) Write(p []byte) (int, error) {
	if bs.pos >= len(bs.data) {
		return 0, io.EOF
	}
	n := copy(bs.data[bs.pos:], p)
	bs.pos += n
	return n, nil
}

func (bs *ByteSlice) Seek(offset int, whence int) (int, error) {
	var newPos int
	switch whence {
	case 0: // io.SeekStart
		newPos = offset
	case 1: // io.SeekCurrent
		newPos = bs.pos + offset
	case 2: // io.SeekEnd
		newPos = len(bs.data) + offset
	default:
		return 0, fmt.Errorf("invalid whence")
	}
	if newPos < 0 || newPos > len(bs.data) {
		return 0, fmt.Errorf("seek out of bounds")
	}
	bs.pos = newPos
	return bs.pos, nil
}
