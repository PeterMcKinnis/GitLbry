package glib

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/pkg/errors"
)

// Creates a new repo on the lbry network
func CliInit(lbryUrl string) error {

	// Resolve repo channel
	c, err := resolveNewStream(lbryUrl);
	if err != nil {
		return err;
	}

	// Don't re-create if repo already exists
	_, err = lbryResolve(lbryUrl)
	if err == nil {
		return errors.New("a repo with the given name already exists")
	}

	// Add self as an author
	config := loadConfig();
	user := config.Default.PushAs;
	if user == nil {
		return errors.New("Cannot create repo, please set the current user with the command\ngitlbry me <channel_url>");
	}

	repo := glSettings{
		Gitlbry: 1,
		Deleted: []string{},
		Authors: []*glAuthor{
			{
				ClaimId:     user.ClaimId,
				ChannelName: user.Name,
				Times:       []int64{time.Now().Unix()},
			},
		},
	}

	repoBytes, err := json.Marshal(repo)
	if err != nil {
		return err
	}

	// Make a unique file name
	tempPath, err := newTempPath()
	if err != nil {
		return err
	}

	// Write temp file to disk
	err = os.WriteFile(tempPath, repoBytes, 0666)
	if err != nil {
		return err
	}

	// Send temp file to lbry network
	if c.channel == nil {
		err = lbryStreamCreate(c.streamName, "0.001", tempPath)
	} else {
		err = lbryStreamCreateOnChannel(c.streamName, c.channel.claimId, "0.001", tempPath);
	}
	if err != nil {
		return err;
	}

	fmt.Println("created")
	return nil
}

// Creates a new repo on the lbry network
func CliAuthorList(lbryUrl string) error {
	
	path, err := newTempPath();
	if err != nil {
		return err;
	}

	r, err := downloadSettings(lbryUrl, path)
	if err != nil {
		return err
	}

	now := time.Now().Unix()
	for _, author := range r.Authors {
		status := "revoked"
		if hasPermission(author, now) {
			status = "granted"
		}
		fmt.Printf("%v:%v %v\n", author.ChannelName, author.ClaimId, status)
	}

	return nil

}

// Creates a new repo on the lbry network
func CliAuthorModify(lbryUrl string, prefixedChannelUrl []string) error {

	claim, err := resolveStream(lbryUrl)
	if err != nil {
		return	errors.Wrapf(err, "error resolving %v.  The url may be malformed or may not reference a git repo");
	}
	if !claim.isMine {
		return errors.New("you do not have permissions to modify the authors")
	}

	path, err := newTempPath();
	if err != nil {
		return err;
	}
	
	settings, err := downloadSettings(lbryUrl, path)
	if err != nil {
		return err
	}

	now := time.Now().Unix()
	for _, x := range prefixedChannelUrl {

		revoke := strings.HasPrefix(x, "^");
		url := x;
		if revoke {
			url = url[1:]
		}

		ch, err := resolveChannel(url)
		if err != nil {
			return errors.Wrapf(err, "error resolving channel %v.", url)
		}

		if revoke {
			settings.revoke(ch.name, ch.claimId, now);
		} else {
			settings.grant(ch.name, ch.claimId, now);
		}

	}

	err = saveRepo(*claim, settings);
	if err != nil {
		return err
	}

	fmt.Println("ok");
	return nil;
}

func CliMeShow() error {
	config := loadConfig();
	me := config.Default.PushAs;
	if me == nil {
		fmt.Print("<current author not set>\n")
		return nil
	}

	fmt.Printf("%v:%v\n", me.Name, me.ClaimId);
	return nil;
}

func CliMeSet(channelUrl string) error {
	
	ch, err := resolveChannel(channelUrl);
	if err != nil {
		return err;
	}

	// Ensure that the channel url is owned by current user
	if !ch.isMine {
		return errors.Errorf("cannot publish as %v:%v, you do not own this channel\n", ch.name, ch.claimId);
	}

	// Load config
	config := loadConfig();

	// Update config
	config.Default.PushAs = &glChannel{
		ClaimId: ch.claimId,
		Name: ch.name,
	}

	// Save config
	err = config.save();
	if err != nil {
		return err;
	}

	fmt.Printf("ok\n");
	return nil;
}





func hasPermission(author *glAuthor, time int64) bool {

	n := len(author.Times)

	i := 0
	for ; i+1 < n; i += 2 {
		start := author.Times[i]
		end := author.Times[i+1]
		if start <= time && time < end {
			return true
		}
	}

	if i == n-1 {
		start := author.Times[i]
		return start <= time
	}

	return false

}


func saveRepo(claim claim, repo *glSettings) error {

	repoBytes, err := json.Marshal(repo)
	if err != nil {
		return err
	}

	// Make a unique file name
	tempPath, err := newTempPath()
	if err != nil {
		return err
	}

	// Write temp file to disk
	err = os.WriteFile(tempPath, repoBytes, 0666)
	if err != nil {
		return err
	}

	// Send temp file to lbry network
	return lbryStreamUpdate(claim.claimId, tempPath)

}

type claim struct {
	url string
	name string
	claimId string
	isMine bool
}

type newStreamClaim struct {
	channel *claim
	streamName string
}

// Nice url is a channel url but is allowed to be missing the "lbry://" or "lbry://@" prefix
func resolveChannel(niceUrl string) (*claim, error) {

	// Make into a full url
	url := prefixNiceChannel(niceUrl);

	// Verify Url
	u, err := NewLbryUrl(url);
	if err != nil {
		return nil, err;
	}
	_, hasStream := u.StreamName();
	if hasStream {
		return nil, errors.New("Expected Channel Url, got Stream Url");
	}

	// Resolve on lbry network
	c, err := lbryResolve(url);
	if err != nil {
		return nil, err;
	}

	if err != nil {
		return nil, err;
	}

	return &claim{
		url: url,
		name: c.NormalizedName,
		claimId: c.ClaimId,
		isMine: *c.IsMyOutput,
	}, nil;

}

// Used to get information about the claim for the channel of a strea
// returns nil if the url doesn't have a channel listed
func resolveNewStream(niceUrl string) (*newStreamClaim, error) {

	// Make into a full url
	url := prefixNice(niceUrl);

	// Verify Url
	u, err := NewLbryUrl(url);
	if err != nil {
		return nil, err;
	}
	streamName, hasStream := u.StreamName();
	if !hasStream {
		return nil, errors.New("expected sream url, got channel url");
	}

	// Get Url for just the channel
	chName, hasChannel := u.ChannelWithModifiers();
	if !hasChannel {
		return &newStreamClaim{
			streamName: streamName,
		}, nil;
	}
	channeUrl := "lbry://@" + chName;


	// Resolve on lbry network
	ch, err := lbryResolve(channeUrl);
	if err != nil {
		return nil, err;
	}

	if err != nil {
		return nil, err;
	}

	return &newStreamClaim{
		streamName: streamName,
		channel: &claim{
			url: channeUrl,
			name: ch.NormalizedName,
			claimId: ch.ClaimId,
			isMine: *ch.IsMyOutput,
		},
	}, nil;

}

func resolveStream(niceUrl string) (*claim, error) {

	// Make into a full url
	url := prefixNice(niceUrl);

	// Verify Url
	u, err := NewLbryUrl(url);
	if err != nil {
		return nil, err;
	}
	_, isStream := u.StreamName();
	if !isStream {
		return nil, errors.New("expected stream url, got channel url")
	}

	// Resolve on lbry network
	c, err := lbryResolve(url);
	if err != nil {
		return nil, err;
	}

	if err != nil {
		return nil, err;
	}

	return &claim{
		url: url,
		name: c.NormalizedName,
		claimId: c.ClaimId,
		isMine: *c.IsMyOutput,
	}, nil;

}

func prefixNiceChannel(niceUrl string) string {

	if strings.HasPrefix(niceUrl, "lbry://@") {
		return niceUrl
	};

	if strings.HasPrefix(niceUrl, "@") {
		return "lbry://" + niceUrl;
	}
		
	return "lbry://@" + niceUrl
}

func prefixNice(niceUrl string) string {

	if strings.HasPrefix(niceUrl, "lbry://") {
		return niceUrl
	};

	return "lbry://" + niceUrl
}



/*
// Returns the streamId for a repo owned by the current user, returns an error
// if the stream does not exist, or is owned by another user.
func myStreamId(lbryUrl string) (string, error) {

	c, err := lbryResolve(lbryUrl)
	if err != nil {
		return "", err
	}

	if c.IsMyOutput == nil {
		// this should never happen as lbryResolve always ask for IsMyOutput
		return "", errors.New("Internal error, could not determine ownership of claim")
	}

	if !*c.IsMyOutput {
		return "", errors.Errorf("%v is not owned by me", lbryUrl)
	}

	return c.ClaimId, nil

}


// Looks up the claim id for a channel owned by the current user
// or returns an error if the channel cannot be found.
func lbryMyChannelClaimId(url lbryUrl) (claimId string, er error) {

	type arg struct {
		Name string `json:"name"`
	}

	type out struct {
		ClaimId        string `json:"claim_id"`
		NormalizedName string `json:"normalized_name"`
	}

	r, err := rpcCall[arg, sdkPage[out]]("channel_list", arg{
		Name: channelName,
	})

	if err != nil {
		return "", "", err
	}

	if r.TotalItems == 0 {
		return "", "", errors.Errorf("Channel with name %v not found", channelName)
	}

	if r.TotalItems > 1 {
		return "", "", errors.Errorf("Found %v channels with name %v", r.TotalItems, channelName)
	}

	return r.Items[0].ClaimId, r.Items[0].NormalizedName, nil

}

// Looks up all channels owned by the current user
func myChannels() ([]*glChannel, error) {

	type arg struct {
		PageSize int `json:"page_size"`
	}

	type out struct {
		ClaimId        string `json:"claim_id"`
		NormalizedName string `json:"normalized_name"`
	}

	r, err := rpcCall[arg, sdkPage[out]]("channel_list", arg{
		PageSize: 5000,
	})

	if err != nil {
		return nil, err;
	}

	// TODO handle multiple pages

	var result []*glChannel
	for _, item := range r.Items {
		result = append(result, &glChannel{
			Name: item.NormalizedName,
			ClaimId: item.ClaimId,
		})
	}

	return result, nil;

}

*/
