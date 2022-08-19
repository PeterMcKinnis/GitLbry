package main

import (
	"fmt"
	"log"
	"os"
	"strings"
)

func Main() (er error) {

	// Parse command line args
  if len(os.Args) > 3 {
    return fmt.Errorf("Usage: git-remote-lbry remote-name url")
  }
	OutPrintf(fmt.Sprintf("args, %v", os.Args));

	// Startup
  lbryUrl := os.Args[2]
	s, err := startup(lbryUrl);
	if err != nil {
		return err;
	}

	for {
		// Note that command will include the trailing newline.
		command, err := readLine()
		if err != nil {
			return err
		}
		switch  {
		case strings.HasPrefix(command, "capabilities"):
			capabilities();
		case strings.HasPrefix(command, "list for-push"):
			s.listForPush();
		case strings.HasPrefix(command, "list"):
			s.list();
		case strings.HasPrefix(command, "fetch"):
			s.fetch(command);
		case strings.HasPrefix(command, "push"):
			s.push(command);
		case command == "":
			return nil
		default:
			return fmt.Errorf("Received unknown command %q", command)
		}
	}
}

func main() {
  if err := Main(); err != nil {
    log.Fatal(err)
  }
}


