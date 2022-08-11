package main

import (
	"bufio"
	"errors"
	"fmt"
	"strconv"
)

const charSpace byte = 32
const char1 byte = 49

type gitObjHeader struct {
	type_  string
	length int64
}

// Same as bufio.Reader.ReadString but excludes the delimter from the result
func ReadString(r bufio.Reader, delim byte) (string, error) {
	x, err := r.ReadString(delim)
	if err != nil {
		return "", err
	}
	return x[:len(x)-1], nil
}

func ReadGitObjHeader(r bufio.Reader) (gitObjHeader, error) {
	type_, err := ReadString(r, charSpace)
	if err != nil {
		return gitObjHeader{}, err
	}

	lenStr, err := ReadString(r, 0)
	if err != nil {
		return gitObjHeader{}, err
	}

	len, err := strconv.ParseInt(lenStr, 10, 64)
	if err != nil {
		return gitObjHeader{}, err
	}

	return gitObjHeader{
		type_:  type_,
		length: len,
	}, nil
}

func readSha1(r bufio.Reader) (hash, error) {
	var result [20]byte
	_, err := r.Read(result[:])
	if err != nil {
		var zero hash
		return zero, err
	}
	return hash(result), nil
}

type GitTreeChild struct {
	mode string
	name string
	sha  hash

	// True if the referenced object is a blob, false if it is a tree
	isBlob bool
}

func readExpected(r bufio.Reader, expected string) error {
	str, err := ReadString(r, charSpace)
	if err != nil {
		return err
	}
	if str != expected {
		return errors.New(fmt.Sprintf("bad git object format, expected %v got %v", expected, str))
	}
	return nil
}

func readTreeChild(r bufio.Reader) (GitTreeChild, int64, error) {

	mode, err := ReadString(r, charSpace)
	if err != nil {
		return GitTreeChild{}, 0, err
	}

	name, err := ReadString(r, 0)
	if err != nil {
		return GitTreeChild{}, 0, err
	}

	sha, err := readSha1(r)
	if err != nil {
		return GitTreeChild{}, 0, err
	}

	isBlob := len(mode) == 6 && mode[0] == char1
	n := int64(len(mode)) + int64(len(name)) + 22

	return GitTreeChild{
		mode:   mode,
		name:   name,
		sha:    sha,
		isBlob: isBlob,
	}, n, nil
}

// Get all objects pointed to by object with content in the given
// stream.
func getDirectRefs(r bufio.Reader) ([]hash, error) {

	hdr, err := ReadGitObjHeader(r)
	if err != nil {
		return nil, err
	}

	switch hdr.type_ {
	case "blob":
		// Blobs are leafs, return empty array
		var out []hash
		return out, nil
	case "tree":
		// Tree
		var nLeft = hdr.length
		var out []hash
		for nLeft > 0 {
			entity, n, err := readTreeChild(r)
			if err != nil {
				return nil, err
			}
			out = append(out, entity.sha)
			nLeft -= n
		}
		return out, nil
	case "tag":
		// For Tag object format refer to
		// https://stackoverflow.com/questions/10986615/what-is-the-format-of-a-git-tag-object-and-how-to-calculate-its-sha
		err := readExpected(r, "object")
		sha, err := readSha1(r)
		if err != nil {
			return nil, err
		}

		out := []hash{sha}
		return out, err

	case "commit":

		// To Commit object format refer to
		// https://stackoverflow.com/questions/22968856/what-is-the-file-format-of-a-git-commit-object-data-structure

		var out []hash

		// Read the sha of the root directory for the commit
		err := readExpected(r, "tree")
		sha, err := readSha1(r)
		if err != nil {
			return nil, err
		}
		out = append(out, sha)

		// Read each parent commit for the commit (usually this has 1 parent.
		// But there can be zero for the first commit, or two for merges
		for {
			parent, err := ReadString(r, charSpace)
			if err != nil {
				return nil, err
			}

			if parent != "parent" {
				break
			}

			sha, err := readSha1(r)
			if err != nil {
				return nil, err
			}
			out = append(out, sha)
			return out, nil
		}
		return out, nil
	default:
		return nil, errors.New("unknown git object type")
	}
}

/// Finds and adds all objects that
func getAllRefs(r Repo, sha hash, results map[hash]struct{}) error {

	// Add the sha hash to the results
	results[sha] = struct{}{}

	// Get a reader for the contents of the object with the
	// given sha hash
	reader, err := r.getObjectReader(sha)
	if err != nil {
		return err
	}

	// Find the direct refs for the given object
	children, err := getDirectRefs(reader)
	if err != nil {
		return err
	}

	// Recursively vist each ref
	for _, child := range children {

		// This prevents infinit recursion if there is a cyclic
		// graph for some reason
		_, ok := results[child]
		if !ok {
			continue
		}

		err := getAllRefs(r, child, results)
		if err != nil {
			return err
		}
	}

	return nil
}

// Finds all objects that are referenced directly or indirectly
// by the object who's hash is sha1
func GitObjectRefTree(r Repo, sha hash) ([]hash, error) {

	// Recursively visit each node
	a := make(map[hash]struct{})
	err := getAllRefs(r, sha, a)

	// Convert set to slice
	var b []hash
	for value := range a {
		b = append(b, value)
	}

	return b, err
}
