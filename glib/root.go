package glib

import "fmt"

type glSettings struct {

	// Version of gitlbry repo.  1 is the only supported value
	Gitlbry int         `json:"gitlbry"`
	Authors []*glAuthor `json:"authors"`
	Deleted []string    `jsion:"deleted"`
}

func (s *glSettings) grant(name string, channelId string, time int64) {

	for _, author := range s.Authors {
		if author.ClaimId == channelId {

			n := len(author.Times)
			odd := n%2 == 1
			if odd {
				// Already have permissions
				return
			}

			author.Times = append(author.Times, time)
			return
		}
	}

	// Add author to list
	s.Authors = append(s.Authors, &glAuthor{
		ClaimId:     channelId,
		ChannelName: name,
		Times: []int64{
			time,
		},
	})
}

func (s *glSettings) revoke(name string, channelId string, time int64) {

	fmt.Printf("revoking from %v %v\n", name, channelId)
	for _, author := range s.Authors {
		if author.ClaimId == channelId {

			n := len(author.Times)
			even := n%2 == 0
			if even {
				// Already revoked
				return
			}

			author.Times = append(author.Times, time)
			return
		}
	}
}

type glRange struct {

	// minimum patch index inclusive
	Start int64 `json:"start"`

	// maximum patch index exclusive, or -1 to indicate no maximum
	End int64 `json:"end"`
}

type glAuthor struct {

	// The lbry claim_id for the author's channel.  e.g. "e66aa0b46d98caf5aeafcee0bbb89bdafec0de72"
	ClaimId string `json:"claim_id"`

	// The authors lbry channel e.g. "gitlbry"
	ChannelName string `json:"channel_name"`

	// These contain a list of increasing time stamps.  Times are in seconds from the unix epoch.
	//  The first time stamp allows the channel push access after the gvien time
	// the next entry disallows after the given time, the thrird re-allows asfter the gvien time etc.
	Times []int64 `json:"ranges"`
}
