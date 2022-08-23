package glib

import (
	"errors"
	"fmt"
	"os/exec"
	"regexp"
)

type PushData struct {
	// This is the raw command sent by git on the command line
	// e.g. "push /refs/heads/master:/refes/heads/master"
	raw string

	// True if git wants us to force the push
	force bool

	// the name of the ref on the local repo
	// e.g. "/refs/heads/master".  commonly src and dst are the same
	src string

	// the name of the ref on the remote repo
	// e.g. "/refs/heads/master".  commonly src and dst are the same
	dst string
}

var PushArgRe = regexp.MustCompile(`^push (\+?)([^:]*):([^:]*)$`)

func parsePushLine(raw string) (PushData, error) {

	matches := PushArgRe.FindStringSubmatch(raw)
	if len(matches) != 4 {
		return zero[PushData](), errors.New("error parsing push command")
	}

	return PushData{
		raw:   raw[5:],
		force: matches[1] == "+",
		src:   matches[2],
		dst:   matches[3],
	}, nil
}

func (s Startup) push(firstLine string) error {

	// Get Channel to push with
	cfg := loadConfig()
	if cfg.Default.PushAs == nil {
		return errors.New("before pushing need to set author.  See gitlbry me <lbry_channel>\n")
	}
	authrorId := cfg.Default.PushAs.ClaimId

	// Parse all grouped pushes from standard in
	OutPrintf("reading push commands")
	args, err := s.readPushCommnds(firstLine)
	if err != nil {
		return err
	}

	//  Attemp to push locally to file://.gitlbry/<repohash>/.git
	OutPrintf("attempting to push locally to ./.gitlbry/<repohash>/")
	err = s.pushAllLocal(args)
	if err != nil {
		writePushResultError(args)
		return err
	}

	// Pack Objects
	OutPrintf("packing objects")
	err = s.createBundle(authrorId, args)
	if err != nil {
		writePushResultError(args)
		return err
	}

	// Done
	OutPrintf("push success")
	writePushResultOk(args)
	return nil

}

// Reads and parses push commands from stdin until a plank line
// is encountered.  Consumes the blank line
func (s Startup) readPushCommnds(firstLine string) ([]PushData, error) {
	lines := []string{firstLine}

	for {
		line, err := readLine()
		if err != nil {
			return nil, err
		}
		if line == "" {
			break
		}
		lines = append(lines, line)
	}

	var out []PushData
	for _, line := range lines {
		arg, err := parsePushLine(line)
		if err != nil {
			return nil, err
		}
		out = append(out, arg)
	}

	return out, nil
}

func (s Startup) pushAllLocal(args []PushData) error {

	for _, arg := range args {

		// Todo get sha hash for arg.src
		cmdArgs := []string{
			"push",
			s.rh.gitRemoteClonePath(),
			arg.raw,
		}
		out, err := exec.Command("git", cmdArgs...).CombinedOutput()
		OutPrintf("git %v %v", cmdArgs, string(out))

		if err != nil {
			return err
		}
	}
	return nil

}

func (s Startup) createBundle(authorId string, pd []PushData) error {

	// Construct command line args
	filePath := s.rh.outBundlePath(s.sync.Index)
	cmdArgs := []string{
		"bundle",
		"create",
		filePath,
	}

	// Include all objects being pushed
	for _, include := range pd {
		cmdArgs = append(cmdArgs, include.dst)
	}

	// Exclude everything that was pushed in a prior bundle
	// May want to get rid of this to have full, self contained
	// bundles??
	for _, exclude := range s.refs {
		cmdArgs = append(cmdArgs, "^"+exclude.ref.toHexString())
	}

	// Have git create the bundle
	out, err := exec.Command("git", cmdArgs...).CombinedOutput()
	OutPrintf("git %v %v", cmdArgs, string(out))

	// Upload to lbry
	name := fmt.Sprintf("%v-%v", s.rh.name, s.sync.DownloadIndex)
	description := s.sync.DownloadPriorHash
	err = lbryStreamCreateForBundle(name, authorId, description, "0.001", filePath)
	if err != nil {
		return err
	}

	return err
}

func writePushResultOk(x []PushData) {
	OutPrintf("Writing Push Results Success len: %v", len(x))
	for _, a := range x {
		Printf("ok %v\n", a.dst)
	}
	Printf("\n")
}

func writePushResultError(x []PushData) {
	OutPrintf("Writing Push Results Error len: %v", len(x))
	for _, a := range x {
		Printf("error %v\n", a.dst)
	}
	Printf("\n")
}
