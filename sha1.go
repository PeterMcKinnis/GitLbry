package main

import "encoding/hex"

type hash [20]byte

func (s hash) toHexString() string {
	return hex.EncodeToString(s[:])
}
