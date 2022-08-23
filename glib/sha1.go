package glib

import (
	"encoding/hex"
	"errors"
)

type NamedRef struct {
	name string
	ref  Sha
}

type Sha [20]byte

func (s Sha) toHexString() string {
	return hex.EncodeToString(s[:])
}

func ShaFromHexString(s string) (Sha, error) {

	x, err := hex.DecodeString(s)
	if err != nil {
		return zero[Sha](), err
	}

	if len(x) != 20 {
		return zero[Sha](), errors.New("error decoding sha hash, bad length")
	}

	var result Sha
	for i := 0; i < 20; i += 1 {
		result[i] = x[i]
	}

	return result, nil

}

func (s Sha) isRef() {}
