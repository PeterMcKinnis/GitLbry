package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
)

var stdinReader = bufio.NewReader(os.Stdin)
var verbose bool = true

func readLine() (string, error) {
	line, err := stdinReader.ReadString('\n')
	if err != nil {
		return "", err
	}

	InPrintf(line)

	line = strings.TrimSuffix(line, "\n")
	return line, nil
}

// Writes output to standard out, will be echoed to
// stderror if verbose
func Printf(format string, s ...any) {
	// Echo to std out for debugging
	OutPrintf(format, s...)
	fmt.Printf(format, s...)
}

// Writes message to stderr for diagnostics and debugging
// Output will start with ">>"
func OutPrintf(format string, s ...any) {
	if verbose {
		format = ">> " + format
		msg := fmt.Sprintf(format, s...)
		if !strings.HasSuffix(msg, "\n") {
			msg = msg + "\n"
		}
		fmt.Fprint(os.Stderr, msg)
	}
}

// Writes message to stderr for diagnostics and debugging
// Output will start with "<<"
func InPrintf(format string, s ...any) {
	if verbose {
		format = "<< " + format
		msg := fmt.Sprintf(format, s...)
		if !strings.HasSuffix(msg, "\n") {
			msg = msg + "\n"
		}
		fmt.Fprint(os.Stderr, msg)
	}
}

func copyFile(src string, dst string) error {

	stat, err := os.Stat(src)
	if err != nil {
		return err
	}

	if !stat.Mode().IsRegular() {
		return errors.New("not a regular file")
	}

	src_, err := os.Open(src)
	if err != nil {
		return err
	}
	defer src_.Close()

	dst_, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dst_.Close()

	_, err = io.Copy(dst_, src_)
	return err
}
