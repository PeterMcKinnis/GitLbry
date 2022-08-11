package main

import "bufio"

type Repo interface {
	readObject(sha hash) (bufio.Reader, error)
	writeObject([]byte) error

	getRefs() ([]Ref, error)
	getHead() (SymbolicRef, error)
}

type Ref struct {
	name string
	sha hash
}

type SymbolicRef struct {
	// The string HEAD
	name string

	// If head is a symbolic ref, this is the path to the symbolic ref
	ref *string

	// The hash of the object pointed to by ref
	hash hash
}