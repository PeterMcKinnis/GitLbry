package glib

import (
	"errors"
	"strings"
)

type OptionArg struct {
	name  string
	value string
}

func parseOptionArg(raw string) (OptionArg, error) {

	matches := strings.SplitN(raw, " ", 3)
	if len(matches) != 3 {
		OutPrintf("failed to parse option: %v", raw)
		return zero[OptionArg](), errors.New("error parsing option command")
	}

	return OptionArg{
		name:  matches[1],
		value: matches[2],
	}, nil
}

func option(command string) error {

	arg, err := parseOptionArg(command)
	if err != nil {
		return err
	}

	OutPrintf("setting option %v to %v", arg.name, arg.value)

	if arg.name == "verbosity" {
		if arg.value != "1" {
			verbose = true
		}
		Printf("ok\n")
	} else {
		Printf("unsupported\n")
	}

	return nil
}
