package main

import (
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"

	"golang.org/x/text/unicode/norm"
)


type RepoHash string 

func (rh RepoHash) inInfoPath() string {
	return fmt.Sprintf(".gitlbry/%s/in/info.json", rh);
}

func (rh RepoHash) inBundlePath(index int) string {
	return fmt.Sprintf(".gitlbry/%s/in/%d.bundle", rh, index);
}

func (rh RepoHash) outBundlePath(index int) string {
	return fmt.Sprintf(".gitlbry/%s/out/%d.bundle", rh, index);
}

func (rh RepoHash) nextOutBundlePath() (string, error) {
	
	OutPrintf("finding next bundle path")
	n := 0;
	for {
		path := rh.outBundlePath(n);
		exists, err := fileExists(path);
		if (err != nil) {
			return "", nil;
		}
		if !exists {
			OutPrintf("next bundle path is %v", path);
			return path, nil;
		}
		n += 1;
	} 
}

func (rh RepoHash) rootPath() string {
	return fmt.Sprintf(".gitlbry/%s", rh);
}

func (rh RepoHash) inPath() string {
	return fmt.Sprintf(".gitlbry/%s/in", rh);
}

func (rh RepoHash) outPath() string {
	return fmt.Sprintf(".gitlbry/%s/out", rh);
}

func (rh RepoHash) gitRemoteClonePath() string {
	return fmt.Sprintf(".gitlbry/%s", rh);
}

func (rh RepoHash) headPath() string {
	return fmt.Sprintf(".gitlbry/%s/.git/HEAD", rh);
}


func NewRepoHash(lbryUrl string) RepoHash {

	form := norm.NFD;
	a := form.String(lbryUrl);
	b := strings.ToLower(a);
	c := sha1.Sum([]byte(b));
	d := hex.EncodeToString(c[:]);
	OutPrintf("Created Repo hash from %v\n   %v", lbryUrl, d);
	return RepoHash(d);
}

type Startup struct {
	rh RepoHash
	refs []NamedRef

	// The value of the head.  This could be a symbolic ref e.g.
	// "@refs/heads/master" or the sha1 hash of a commit e.g. "ccdddd6c5b19436e52146dfc11fd8632ca60b31b" 
	head string
} 

type InInfo struct {
	Index int `json:"bundleIndex"`;
}

func zero[T any]() T {
	var x T;
	return x;
}

func startup(lbryurl string) (Startup, error) {

	rh := NewRepoHash(lbryurl);

	// Aquire filesystem lock
	err := rh.lock();
	OutPrintf("Aquireing lock")
	if err != nil {
		return zero[Startup](), err;
	}

	// Initialize if necessary
	OutPrintf("Initializing")
	err = rh.initialize();
	if err != nil {
		return zero[Startup](), err;
	}

	// Load info.json from disk
	OutPrintf("loading in-info");
	oldInfo, err := rh.loadInInfo()
	if err != nil {
		return zero[Startup](), err;
	}

  // Update .gitlbry/<reposhash>/in from the lbry network
	OutPrintf("getting changes from lbry network (mocked for now)");
	newInfo, err := rh.downloadChanges(oldInfo)
	if err != nil {
		return zero[Startup](), err;
	}

	// Apply changes to a local .git repo that clones lbry
	OutPrintf("applying remote changes to a local clone");
	err = rh.applyChanges(oldInfo.Index, newInfo.Index);
	if err != nil {
		return zero[Startup](), err;
	}

	// Update info.json
	OutPrintf("updateing in-info");
	err = rh.saveInInfo(newInfo);
	if err != nil {
		return zero[Startup](), err;
	}

	// Load Regular references
	OutPrintf("loading references");
	refs, err := rh.loadRefs();
	if err != nil {
		return zero[Startup](), err;
	}

	// Load head ref
	OutPrintf("loading HEAD reference");
	head, err := rh.loadHead();
	if err != nil {
		return zero[Startup](), err;
	}

	// Done, success
	OutPrintf("startup success");
	return Startup{
		rh:  rh,
		refs: refs,
		head: head,
	}, nil;

}


func (rh RepoHash) loadRefs() ([]NamedRef, error) {
	
	// Run git command show-ref
	cmd :=	exec.Command(
		"git",
		"show-ref",
	);
	cmd.Dir = rh.gitRemoteClonePath();
	out, err := cmd.CombinedOutput()
	OutPrintf("git show-ref %s", out);

	// Special case for newly created repo, show ref exits with status 1
	// and out is nil
	if len(out) == 0 {
		return nil, nil;
	} 
	if err != nil {
		return nil, err;
	}

	// Parse a named ref from a line of the outputt 
	parseNamedRef := func (x string) (NamedRef, error) {
		subs := strings.SplitN(x, " ", 2);
		if (len(subs) != 2) {
			return zero[NamedRef](), errors.New("unexpected format while parsing output of git show-ref");
		}

		ref, err := ShaFromHexString(subs[0]);
		if err != nil {
			return zero[NamedRef](),err;
		}

		return NamedRef{
			name: subs[1],
			ref: ref,
		}, nil;
	}

	// Parse output
	o := string(out);
	o = strings.TrimRight(o, "\n");
	lines := strings.Split(o, "\n");
	var results []NamedRef; 
	for _, line := range lines {
		nr, err := parseNamedRef(line);
		if err != nil {
			return nil, err;
		}
		results = append(results, nr);
	}

	return results, nil;
}

// Loads head.  If head a symbolic ref will be formated prefixed with
// @ e.g.
// "@ref/heads/master" otherswise will be hex encoded sha1 hash
// e.g. "ccdddd6c5b19436e52146dfc11fd8632ca60b31b"
func (rh RepoHash) loadHead() (string, error) {
	b, err := os.ReadFile(rh.headPath());
	if err != nil {
		return "", err;
	}
	out := strings.Replace(string(b), "ref: ", "@", 1);
	out = strings.TrimRight(out, "\n");
	return out, nil;
}

func (rh RepoHash) loadInInfo() (InInfo, error) {
	path := rh.inInfoPath();
	b, _ := ioutil.ReadFile(path);
	var i InInfo
	err := json.Unmarshal(b, &i)
	return i, err;
}

func (repoHash RepoHash) saveInInfo(info InInfo) (error) {
	b, err := json.Marshal(info);
	if err != nil {
		return err;
	}
	path := repoHash.inInfoPath();
	return os.WriteFile(path, b, 0666);
}

func (rh RepoHash) initialize() error {
	
	// Check if file exists
	exists, err := fileExists(rh.rootPath());
	if err != nil {
		return err;
	}
	if exists {
		OutPrintf("Initializing done - already initialized\n")
		// Already initialized
		return nil;
	}

	// Initialize
	OutPrintf("Createing Root Directory\n")
	err = os.MkdirAll(rh.rootPath(), 0777);
	if err != nil {
		return err;
	}
	
	OutPrintf("Createing Out Directory\n")
	os.MkdirAll(rh.outPath(), 0777);
	if err != nil {
		return err;
	}

	OutPrintf("Createing In Directory\n")
	os.MkdirAll(rh.inPath(), 0777);
	if err != nil {
		return err;
	}

	OutPrintf("Createing InInfo\n")
	var info InInfo;
	err = rh.saveInInfo(info);
	if err != nil {
		return err;
	}

	OutPrintf("Running Git Init\n")
	cmd := exec.Command("git", "init", "--bare");
	cmd.Dir = rh.gitRemoteClonePath();
	
	out, err := cmd.CombinedOutput();
	OutPrintf("%s", out);
	return err;

}

func (rh RepoHash) downloadChanges(info InInfo) (InInfo, error) {

	n := info.Index;
	for {
		
		err := os.Rename(rh.outBundlePath(n), rh.inBundlePath(n))

		// Special case, no more data to download
		if os.IsNotExist(err) {
			break;
		}

		// Fail on fs errors
		if err != nil {
			return zero[InInfo](), err;
		}

		// Prep for next download
		n += 1;

	}

	return InInfo{Index: n}, nil;

}

func (rh RepoHash) applyChanges(start int, end int) error {
	for n := start; n < end; n += 1 {

		cmd :=	exec.Command(
			"git",
			"bundle",
			"unbundle",
			rh.inBundlePath(n),
		);

		cmd.Dir = rh.gitRemoteClonePath();
		
		out, err := cmd.CombinedOutput();
		OutPrintf("git bundle %v", string(out));
		if err != nil {
			return err;
		}
	}

	return nil;
}


func fileExists(path string) (bool, error) {
	OutPrintf("checking if file exists %v", path);
	_, err := os.Stat(path);
	if os.IsNotExist(err) {
		OutPrintf("no it doesn't");
		return false, nil;
	}
	if err == nil {
		OutPrintf("yes it does");
		return true, nil;
	}
	return zero[bool](), err;
}
