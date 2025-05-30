package main

import (
	"log"
	"os"

	"github.com/bugsnag/bugsnag-go"
	"github.com/docker/docker-language-server/internal/pkg/cli"
	"github.com/docker/docker-language-server/internal/pkg/cli/metadata"
)

func main() {
	bugsnag.Configure(bugsnag.Configuration{
		APIKey:     metadata.BugSnagAPIKey,
		AppType:    "languageServer",
		AppVersion: metadata.Version,
		// if it is the empty string it will not be set
		Hostname:        "REDACTED",
		Logger:          log.New(os.Stderr, "", log.LstdFlags),
		ProjectPackages: []string{"main", "github.com/docker/docker-language-server/**"},
		ReleaseStage:    "production",
	})

	cli.Execute()
}
