package main

import (
	"log"
	"os"
	"strings"

	"gitlbry.com/glib"
)

func handleErr(err error) {
	if err != nil {
		log.Fatal(err.Error());
	}
}

func showHelp() {
	log.Fatal(`useage:

	// Create a new repository
	gitlbry init <lbry_url>

	// Get or Set the channel to publish as
	gitlbry me [<channel_url>]

	// View or change permissions.
	gitlbry author <lbry_url> [[^]<channel_url>]*`);
}

func showInitHelp() {
	log.Fatal(`useage:	
gitlbry init <lbry_url>
	
	Creates an empty repo at the given <lbry_url>.  Outputs the perminant lbry
	url.
	
	<lbry_url> The lbry url for the repo. You may omit the "lbry://" prefix for
	           convieniance. 

`);
}

func showMeHelp() {
	log.Fatal(`useage:	
gitlbry me [<channel_url>]
	
  With zero arguements, prints the channel used to push changes
	
	With one arguement, sets the channel used to push changes
	
	<channel_url> The lbry url for the channel.  For convieniance, the prefix
	              "lbry://" or "lbry://@" may be omitted.`);
}

func showAuthorHelp() {
		log.Fatal(`useage:	
gitlbry author <lbry_url> [[^]<channel_url>]*

  With zero <channel_url>, prints a list all channels that have ever had push
  permision along with their current permission.

  With one or more <channel_url> grants and revokes push privilidges for each 
  channel.   Channels prefixed with ^ will privlidge revoked, others will have 
  privledge granted.  Error if any channel cannot be resolved on the lbry 
	network 

  <lbry_url>    A lbry url to the repository.  For convieniance, the prefix 
                "lbry://" may be omitted.

  <channel_url> The lbry url for the channel.  For convieniance, the prefix
	              "lbry://" or "lbry://@" may be omitted.  
`)}


func main() {
	
	args := os.Args[1:]

	if len(args) < 1 {
		showHelp();
	}
	
	// Pop command
	command := strings.ToLower(args[0])
	args = args[1:]

	switch command {
	case "init":
		if len(args) == 1 {
			handleErr(glib.CliInit(args[0]));
			return;
		} else {
			showInitHelp();
		}
	case "me":
		if len(args) == 0 {
			handleErr(glib.CliMeShow());
		} else if len(args) == 1 {
			handleErr(glib.CliMeSet(args[0]));
		} else {
			showMeHelp();
		}
	case "author":
		if len(args) == 1 {
			handleErr(glib.CliAuthorList(args[0]));
		} else if len(args) > 1 {
			handleErr(glib.CliAuthorModify(args[0], args[1:]));
		} else {
			showAuthorHelp();
		}
	default:
		showHelp();
	}
}