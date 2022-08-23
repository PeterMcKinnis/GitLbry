package main

import (
	"log"
	"gitlbry.com/glib"
)

func main() {
	if err := glib.GitRemoteLbry(); err != nil {
		log.Fatal(err)
	}
}
