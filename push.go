package main

import (
	"errors"
	"os/exec"
	"regexp"
)

type PushData struct {
	raw string

	force bool

	src string
	dst string

}

var PushArgRe = regexp.MustCompile(`^push (\+?)([^:]*):([^:]*)$`);

func parsePushLine(raw string) (PushData, error) {

	matches := PushArgRe.FindStringSubmatch(raw);
	if len(matches) != 4 {
		return zero[PushData](), errors.New("error parsing push command");
	}

	return PushData{
		raw: raw[5:],
		force: matches[1] == "+",
		src: matches[2],
		dst: matches[3],
	}, nil;
}


func (s Startup) push(firstLine string) error {

	// Parse all grouped pushes from standard in
	OutPrintf("reading push commands")
	args, err := s.readPushCommnds(firstLine);
	if err != nil {
		return err;
	}

  //  Attemp to push locally to file://.gitlbry/<repohash>/.git 
	OutPrintf("attempting to push locally to ./.gitlbry/<repohash>/")
	err = s.pushAllLocal(args);
	if err != nil {
		writePushResultError(args)
		return err;
	}
	
	// Pack Objects
	OutPrintf("packing objects")
	err = s.packObjects(args);
	if err != nil {
		writePushResultError(args)
		return err;
	}

	// Done
	OutPrintf("push success")
	writePushResultOk(args)
	return nil;

}

// Reads and parses push commands from stdin until a plank line
// is encountered.  Consumes the blank line
func (s Startup) readPushCommnds(firstLine string) ([]PushData, error) {
	lines := []string{firstLine};
	
	for {
		line, err := readLine();
		if err != nil {
			return nil, err;
		}
		if line == "" {
			break;
		}
		lines = append(lines, line);
	}
	
	var out []PushData
	for _, line  := range lines {
		arg, err := parsePushLine(line);
		if err != nil {
			return	 nil, err;
		}
		out = append(out, arg);
	}

	return out, nil;
}


func (s Startup) pushAllLocal(args []PushData) error {

	for _, arg := range args {

		// Todo get sha hash for arg.src
		cmdArgs := []string {
			"push", 
			s.rh.gitRemoteClonePath(), 
			arg.raw,
		};
		out, err := exec.Command("git", cmdArgs...).CombinedOutput();
		OutPrintf("git %v %v", cmdArgs, string(out));

		if (err != nil ) {
			return err;
		}
	}
	return nil;

}

func (s Startup) packObjects(pd []PushData) error {

	cmdArgs :=  []string {
		"bundle", 
		"create", 
	};

	path, err := s.rh.nextOutBundlePath();
	if err!= nil {
		return err;
	}
	cmdArgs = append(cmdArgs, path);

	// Include all objects being pushed
	for _, include := range pd {
		cmdArgs = append(cmdArgs, include.dst)
	}

	// Exclude everything that was pushed in a prior bundle
	for _, exclude := range s.refs {
		cmdArgs = append(cmdArgs, "^" + exclude.ref.toHexString());
	}

	out, err := exec.Command(
		"git", cmdArgs...).CombinedOutput();
	OutPrintf("git %v %v", cmdArgs, string(out));
	return err;
}

func writePushResultOk(x []PushData) {
	OutPrintf("Writing Push Results Success len: %v", len(x))
	for _, a := range x {
		Printf("ok %v\n", a.dst);
	}
	Printf("\n")
}

func writePushResultError(x []PushData) {
	OutPrintf("Writing Push Results Error len: %v", len(x))
	for _, a := range x {
		Printf("error %v\n", a.dst);
	}
	Printf("\n")
}

