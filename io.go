package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)


var stdinReader = bufio.NewReader(os.Stdin);
var verbose bool = true;

func readLine() (string, error) {
	line, err := stdinReader.ReadString('\n')
	if err != nil {
		return "", err
	}

	InPrintf(line);

	line = strings.TrimSuffix(line, "\n");
	return line, nil;
}

// Writes output to standard out, will be echoed to
// stderror if verbose
func Printf(format string, s ...any) {
	// Echo to std out for debugging
	OutPrintf(format, s...);
	fmt.Printf(format, s...);
}


// Writes message to stderr for diagnostics and debugging
// Output will start with ">>"
func OutPrintf(format string, s ...any) {
	if verbose {
		format = ">> " + format;
		msg := fmt.Sprintf(format, s...);
		if !strings.HasSuffix(msg, "\n") {
			msg = msg + "\n";
		}
		fmt.Fprint(os.Stderr, msg);
	}
}

// Writes message to stderr for diagnostics and debugging
// Output will start with "<<"
func InPrintf(format string, s ...any) {
	if verbose {
		format = "<< " + format;
		msg := fmt.Sprintf(format, s...);
		if !strings.HasSuffix(msg, "\n") {
			msg = msg + "\n";
		}
		fmt.Fprint(os.Stderr, msg);
	}
}
