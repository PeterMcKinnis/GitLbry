package main

import (
	"errors"
	"os/exec"
	"regexp"
)



type FetchArg struct {
	name string
	sha Sha
}

var FetchArgRe = regexp.MustCompile(`^fetch ([1-9a-f]*):(.*)$`);

func parseFetchArg(raw string) (FetchArg, error) {

	matches := PushArgRe.FindStringSubmatch(raw);
	if len(matches) != 4 {
		return zero[FetchArg](), errors.New("error parsing fetch command");
	}

	sha, err := ShaFromHexString(matches[0]);
	if err != nil {
		return zero[FetchArg](), err;
	}

	return FetchArg{
		name: matches[1],
		sha: sha,
	}, nil;
}

// Reads and parses push commands from stdin until a plank line
// is encountered.  Consumes the blank line
func (s Startup) readFetchArgs(firstLine string) ([]FetchArg, error) {
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
	
	var out []FetchArg
	for _, line  := range lines {
		arg, err := parseFetchArg(line);
		if err != nil {
			return	 nil, err;
		}
		out = append(out, arg);
	}

	return out, nil;
}

func (s Startup) fetch(firstLine string) error {

	// Parse all grouped pushes from standard in
	args, err := s.readFetchArgs(firstLine);
	if err != nil {
		return err;
	}

  //  Attemp to push locally to file://.gitlbry/<repohash>/.git 
	err = s.fetchPack(args);
	if err != nil {
		return err;
	}
	
	// Write result
	Printf("/n");

	// Done
	return nil;

}

func (s Startup) fetchPack(x []FetchArg) error {

	args := []string {
		"fetch-pack",
		s.rh.gitRemoteClonePath(),
	}

	for _, y := range x {
		args = append(args, y.sha.toHexString());
	}

	err := exec.Command("git", args...).Run();
	Printf("\n");

	return err;

}

