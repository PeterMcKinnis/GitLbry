package main

import (
	"bufio"
	"compress/zlib"
	"crypto/sha1"
	"encoding/hex"
	"io/fs"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
)

var base_dir string = `C:\Users\peter\git-remote-lbry\data`;

type RemoteRepo struct {
	path string 
}

func (rr *RemoteRepo) readObject(sha hash) (*bufio.Reader, error) {

	// Create path to object
	var shaText = sha.toHexString();
	a := shaText[:2];
	b := shaText[2:]
	fullPath := path.Join(base_dir, "objects", a, b);
	
	// Create reader for file
	r, err := os.Open(fullPath);	
	if err != nil {
		return nil, err;
	}
	defer r.Close();

	// Create reader that decompressed data
	r2, err := zlib.NewReader(r)
	if err != nil {
		return nil, err;
	}
	defer r2.Close();

	// Wrap in a bufioReader
	return bufio.NewReader(r2), nil;

}


func (rr *RemoteRepo) writeObject(content []byte) error {

	// Calculate the sha1 hash for the object
	hash := sha1.New().Sum(content)

	// Create path to object
	hashText := hex.EncodeToString(hash)
	a := hashText[:2];
	b := hashText[2:]
	fullPath := path.Join(base_dir, "objects", a, b);

	// Create reader for file
	r, err := os.Create(fullPath);	
	if err != nil {
		return err;
	}
	defer r.Close();

	// Create reader that decompressed data
	r2 := zlib.NewWriter(r)
	defer r2.Close();

	// Write data
	_, err = r2.Write(content);
	return err;

}


func (rr *RemoteRepo) getRefs() ([]Ref, error) {

	refPath := path.Join(base_dir, path, "refs");
	filepath.Walk(refPath, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil;
		}

		content, err := ioutil.ReadFile(path)
		
	});
	

}

func (rr *RemoteRepo) getRefsRecursive(info fs.FileInfo, results []Ref, path []string) error {
	if info.IsDir() {
		files, err := ioutil.ReadDir(info.Sys())
	} else {

	}
}
 

func (rr *RemoteRepo) getHead() (SymbolicRef, error) {

}