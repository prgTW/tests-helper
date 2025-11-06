package main

import (
	"time"

	"github.com/tomaszwojcik/tests-helper/cmd"
)

var (
	version = "snapshot"
	commit  = "<commit-unknown>"
	date    = time.Now().Format(time.RFC3339)
)

func main() {
	cmd.Execute(version, commit, date)
}
