package main

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"strings"
)

func main() {
	if err := cmdMain.Execute(); err != nil {
		os.Exit(1)
	}
}

func findFirstEDID(r io.Reader) (*Edid, error) {
	header, _ := hex.DecodeString("00ffffffffffff00")
	buffer := make([]byte, 16)
	for {
		b := make([]byte, 8)
		n, err := r.Read(b)
		if err != nil {
			return nil, err
		}
		buffer = append(buffer[8:], b[:n]...)
		if idx := bytes.Index(buffer, header); idx >= 0 {
			nextEdid := make([]byte, 128-(16-idx))
			if _, err := io.ReadFull(r, nextEdid); err != nil {
				return nil, err
			}
			return NewEdid(append(buffer[idx:], nextEdid...))
		}
	}
}

func loadEdidFromFile(f *os.File) (*Edid, error) {
	if edid, err := findFirstEDID(f); err == nil {
		return edid, nil
	}
	_, err := f.Seek(0, 0)
	if err != nil {
		return nil, fmt.Errorf("seek to start failed: %v", err)
	}
	stats, err := f.Stat()
	if err != nil {
		return nil, err
	}
	if stats.Size() > 1024 {
		return nil, fmt.Errorf("file size too large")
	}
	strData, err := io.ReadAll(f)
	if err != nil {
		return nil, err
	}
	var hexData string
	for _, v := range []rune(strings.TrimSpace(strings.ReplaceAll(string(strData), "0x", ""))) {
		if v >= '0' && v <= '9' || v >= 'a' && v <= 'f' || v >= 'A' && v <= 'F' {
			hexData += string(v)
		}
	}
	data, err := hex.DecodeString(string(strData))
	if err != nil {
		return nil, err
	}
	if len(data) != 128 {
		return nil, fmt.Errorf("No edid found")
	}
	return NewEdid(data)
}

func applyFirstEDID2NewFile(in, out *os.File, edid *Edid) error {
	_, err := in.Seek(0, 0)
	if err != nil {
		return fmt.Errorf("seek to start failed: %v", err)
	}
	_, err = out.Seek(0, 0)
	if err != nil {
		return fmt.Errorf("seek to start failed: %v", err)
	}
	applyed := false
	header, _ := hex.DecodeString("00ffffffffffff00")
	var buffer = make([]byte, 16)
	for {
		b := make([]byte, 8)
		n, err := in.Read(b)
		if err != nil {
			if err == io.EOF {
				if !applyed {
					return fmt.Errorf("no EDID found")
				}
				return nil
			}
			return err
		}
		if !applyed {
			buffer = append(buffer[8:], b[:n]...)
		}
		if _, err := out.Write(b[:n]); err != nil {
			return err
		}
		if idx := bytes.Index(buffer, header); idx >= 0 && !applyed {
			if _, err := in.Seek(int64(128-(len(buffer)-idx)), 1); err != nil {
				return err
			}
			if _, err := out.Write(edid.ToBytes()[(len(buffer) - idx):]); err != nil {
				return err
			}
			applyed = true
		}

	}
}
