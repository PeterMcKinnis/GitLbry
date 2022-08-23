package glib

import (
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"

	"github.com/pkg/errors"
)

type RepoName struct {

	// The name of the remote according to git
	url lbryUrl

	// The Stram_Name part of the lbry url
	name string

	// Controlls where local files are located
	hash string
}

func (rh RepoName) syncPath() string {
	return fmt.Sprintf("%s/sync.json", rh.rootPath())
}

func (rh RepoName) settingsPath() string {
	return fmt.Sprintf("%s/settings.json", rh.rootPath())
}

func (rh RepoName) inBundlePath(index int) string {
	return fmt.Sprintf("%s/in/%d.bundle", rh, index)
}

func (rh RepoName) outBundlePath(index int) string {
	return fmt.Sprintf("%s/out/%d.bundle", rh, index)
}

func (rh RepoName) inPath() string {
	return fmt.Sprintf("%s/in", rh.rootPath())
}

func (rh RepoName) outPath() string {
	return fmt.Sprintf("%s/out", rh.rootPath())
}

func (rh RepoName) gitRemoteClonePath() string {
	return rh.rootPath()
}

func (rh RepoName) rootPath() string {
	return fmt.Sprintf(".glbry/%s", rh.hash)
}

func (rh RepoName) headPath() string {
	return fmt.Sprintf("%s/.git/HEAD", rh.rootPath())
}

func NewRepoHash(lbryUrl string) (RepoName, error) {

	// Hash
	var h = sha1.Sum([]byte(lbryUrl))
	hash := string(h[:])

	// Url
	u, err := NewLbryUrl(lbryUrl)
	if err != nil {
		return zero[RepoName](), nil
	}

	// name
	name, ok := u.StreamName()
	if !ok {
		return zero[RepoName](), errors.Errorf("invalid url for repo.  Url may refer to a channel instead of a stream")
	}

	return RepoName{
		url:  u,
		hash: hash,
		name: name,
	}, nil

}

type Startup struct {

	// Normalized, perminant path to repo in the form
	// lbry://@<channel_name>#channel_id/path/to/repo
	name string

	rh RepoName

	sync Sync

	// The value of the remote head.  This could be a symbolic ref e.g.
	// "@refs/heads/master" or the sha1 hash of a commit e.g. "ccdddd6c5b19436e52146dfc11fd8632ca60b31b"
	head string

	// A list git refs in the remote (e.g. Tags and Commits)
	refs []NamedRef
}

type Sync struct {
	// The number of bundle files that have been successfully downloaded
	// to the local directory
	DownloadIndex int

	// The sha1 hash of the bundle file at (DownloadIndex-1)
	DownloadPriorHash string

	// The number of bundle files that have been successfully applied
	// to the local repository
	Index int
}

func zero[T any]() T {
	var x T
	return x
}

func startup(lbryurl string) (Startup, error) {

	rh, err := NewRepoHash(lbryurl)
	if err != nil {
		return zero[Startup](), err
	}

	// Aquire filesystem lock
	err = rh.lock()
	OutPrintf("Aquireing lock")
	if err != nil {
		return zero[Startup](), err
	}

	// Initialize Local Directory if necessary
	OutPrintf("Initializing")
	err = rh.initializeLocalDirectory()
	if err != nil {
		return zero[Startup](), err
	}

	// Load sync.json from disk
	OutPrintf("loading sync")
	sync, err := rh.loadSync()
	if err != nil {
		return zero[Startup](), err
	}

	// Download settings from lbry
	OutPrintf("loading settings")
	settings, err := downloadSettings(string(rh.url), rh.settingsPath())
	if err != nil {
		return zero[Startup](), err
	}

	// Update .gitlbry/<reposhash>/in from the lbry network
	OutPrintf("getting changes from lbry network (mocked for now)")
	err = rh.downloadBundles(&sync, settings)
	if err != nil {
		return zero[Startup](), err
	}

	// Apply changes to a local .git repo that clones lbry
	OutPrintf("applying remote changes to a local clone")
	err = rh.applyBundles(sync)
	if err != nil {
		return zero[Startup](), err
	}

	// Update sync.json
	OutPrintf("updateing sync %+v", sync)
	err = rh.saveSync(sync)
	if err != nil {
		return zero[Startup](), err
	}

	// Load Regular references
	OutPrintf("loading references")
	refs, err := rh.loadRefs()
	if err != nil {
		return zero[Startup](), err
	}

	// Load head ref
	OutPrintf("loading HEAD reference")
	head, err := rh.loadHead()
	if err != nil {
		return zero[Startup](), err
	}

	// Done, success
	OutPrintf("startup success")
	return Startup{
		rh:   rh,
		refs: refs,
		head: head,
	}, nil

}

func (rh RepoName) loadRefs() ([]NamedRef, error) {

	// Run git command show-ref
	cmd := exec.Command(
		"git",
		"show-ref",
	)
	cmd.Dir = rh.gitRemoteClonePath()
	out, err := cmd.CombinedOutput()
	OutPrintf("git show-ref %s", out)

	// Special case for newly created repo, show ref exits with status 1
	// and out is nil
	if len(out) == 0 {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	// Parse a named ref from a line of the outputt
	parseNamedRef := func(x string) (NamedRef, error) {
		subs := strings.SplitN(x, " ", 2)
		if len(subs) != 2 {
			return zero[NamedRef](), errors.New("unexpected format while parsing output of git show-ref")
		}

		ref, err := ShaFromHexString(subs[0])
		if err != nil {
			return zero[NamedRef](), err
		}

		return NamedRef{
			name: subs[1],
			ref:  ref,
		}, nil
	}

	// Parse output
	o := string(out)
	o = strings.TrimRight(o, "\n")
	lines := strings.Split(o, "\n")
	var results []NamedRef
	for _, line := range lines {
		nr, err := parseNamedRef(line)
		if err != nil {
			return nil, err
		}
		results = append(results, nr)
	}

	return results, nil
}

// Loads head.  If head a symbolic ref will be formated prefixed with
// @ e.g.
// "@ref/heads/master" otherswise will be hex encoded sha1 hash
// e.g. "ccdddd6c5b19436e52146dfc11fd8632ca60b31b"
func (rh RepoName) loadHead() (string, error) {
	b, err := os.ReadFile(rh.headPath())
	if err != nil {
		return "", err
	}
	out := strings.Replace(string(b), "ref: ", "@", 1)
	out = strings.TrimRight(out, "\n")
	return out, nil
}

func (rh RepoName) loadSync() (Sync, error) {
	path := rh.syncPath()
	b, _ := ioutil.ReadFile(path)
	var i Sync
	err := json.Unmarshal(b, &i)
	return i, err
}

func downloadSettings(lbryUrl string, fileName string) (*glSettings, error) {

	// For Now, just re-download on every invocation
	// Would be better to check the header and only re-download when necessary
	err := lbryGet(lbryUrl, fileName)
	if err != nil {
		return nil, err
	}

	// Open File
	fid, err := os.Open(fileName)
	if err != nil {
		return nil, err
	}
	defer fid.Close()

	// Read and Parse
	var settings glSettings
	err = json.NewDecoder(fid).Decode(&settings)
	if err != nil {
		return nil, err
	}

	return &settings, nil

}

func (repoHash RepoName) saveSync(info Sync) error {
	b, err := json.Marshal(info)
	if err != nil {
		return err
	}
	path := repoHash.syncPath()
	return os.WriteFile(path, b, 0666)
}

func (rh RepoName) initializeLocalDirectory() error {

	// Check if file exists
	exists, err := fileExists(rh.rootPath())
	if err != nil {
		return err
	}
	if exists {
		OutPrintf("Initializing done - already initialized\n")
		// Already initialized
		return nil
	}

	// Initialize
	OutPrintf("Createing Root Directory\n")
	err = os.MkdirAll(rh.rootPath(), 0777)
	if err != nil {
		return err
	}

	OutPrintf("Createing Out Directory\n")
	os.MkdirAll(rh.outPath(), 0777)
	if err != nil {
		return err
	}

	OutPrintf("Createing In Directory\n")
	os.MkdirAll(rh.inPath(), 0777)
	if err != nil {
		return err
	}

	OutPrintf("Creating Sync\n")
	var sync Sync
	err = rh.saveSync(sync)
	if err != nil {
		return err
	}

	OutPrintf("Running Git Init\n")
	cmd := exec.Command("git", "init", "--bare")
	cmd.Dir = rh.gitRemoteClonePath()

	out, err := cmd.CombinedOutput()
	OutPrintf("%s", out)
	return err

}

func (s *glSettings) isDeleted(claimId string) bool {
	for _, deleted := range s.Deleted {
		if deleted == claimId {
			return true
		}
	}
	return false
}

func (s *glSettings) isAuthorized(channelId string, timestamp int64) bool {
	for _, author := range s.Authors {
		if author.ClaimId == channelId {
			i := 0
			for ; i+1 < len(author.Times); i += 2 {
				start := author.Times[i]
				end := author.Times[i+1]
				if start <= timestamp && timestamp < end {
					return true
				}
			}

			if i < len(author.Times) {
				start := author.Times[i]
				if start <= timestamp {
					return true
				}
			}
		}
	}
	return false
}

var BundleNotFoundErr error = errors.New("bundle not found")

func getDescription(raw json.RawMessage) string {
	type arg struct {
		Description string `json:"description"`
	}
	var a arg
	err := json.Unmarshal(raw, &a)
	if err != nil {
		return ""
	}
	return a.Description
}

// Finds the perminant url of bundle with the given name and description
func findBundle(name string, description string, settings *glSettings) (string, error) {

	type arg struct {
		Name       string   `json:"name"`
		ChannelIds []string `json:"channel_ids"`
		PageSize   int      `json:"page_size"`
		OrderBy    string   `json:"height"`
	}

	type signChan struct {
		ClaimId string `json:"claim_id"`
	}

	type out struct {
		withError
		PermanentUrl   string    `json:"permanent_url"`
		ClaimId        string    `json:"claim_id"`
		Timestamp      int64     `json:"timestamp"`
		SigningChannel *signChan `json:"signing_channel"`
		Value          json.RawMessage
	}

	page, err := rpcCall[arg, sdkPage[*out]]("claim_search", arg{
		Name:       name,
		ChannelIds: Map(settings.Authors, func(a *glAuthor) string { return a.ClaimId }),
		PageSize:   5000,
		OrderBy:    "publish_time",
	})

	if err != nil {
		return "", err
	}

	for _, item := range page.Items {
		if item.Error == nil &&
			!settings.isDeleted(item.ClaimId) &&
			item.SigningChannel != nil &&
			settings.isAuthorized(item.SigningChannel.ClaimId, item.Timestamp) &&
			getDescription(item.Value) == description {

			// Found it
			return item.PermanentUrl, nil
		}
	}

	return "", BundleNotFoundErr

}

func (rh RepoName) downloadBundles(sync *Sync, settings *glSettings) error {

	for {

		// Repo Name
		name := fmt.Sprintf("%v-%v", rh.name, sync.DownloadIndex)
		description := sync.DownloadPriorHash

		// Find next bundle
		bundleUrl, err := findBundle(name, description, settings)

		// Check for Successfull completion
		if err == BundleNotFoundErr {
			break
		}
		if err != nil {
			return err
		}

		// Download bundle
		path := rh.inBundlePath(sync.DownloadIndex)
		err = lbryGet(bundleUrl, path)
		if err != nil {
			return err
		}

		// Calc Sha Hash For Next Bundle
		fid, err := os.Open(path)
		if err != nil {
			return err
		}
		hash := sha1.New()
		_, err = io.Copy(hash, fid)
		if err != nil {
			return err
		}
		prior := hex.EncodeToString(hash.Sum(nil))

		// Prep for next download
		sync.DownloadIndex += 1
		sync.DownloadPriorHash = prior
	}

	return nil

}

func (rh RepoName) applyBundles(sync Sync) error {

	for n := sync.Index; n < sync.DownloadIndex; n += 1 {

		cmd := exec.Command(
			"git",
			"bundle",
			"unbundle",
			rh.inBundlePath(n),
		)

		cmd.Dir = rh.rootPath()
		out, err := cmd.CombinedOutput()
		OutPrintf("git bundle %v", string(out))
		if err != nil {
			return err
		}

		sync.Index += 1
	}

	return nil
}

func fileExists(path string) (bool, error) {
	OutPrintf("checking if file exists %v", path)
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		OutPrintf("no it doesn't")
		return false, nil
	}
	if err == nil {
		OutPrintf("yes it does")
		return true, nil
	}
	return zero[bool](), err
}

func Map[T any, U any](items []T, fn func(T) U) []U {
	var result []U
	for _, item := range items {
		result = append(result, fn(item))
	}
	return result
}
