package plugin

import (
	"context"
	"fmt"

	"github.com/drone/drone-go/drone"

	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

func getFilesChanged(r drone.Repo, b drone.Build, token string) ([]string, error) {
	newctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(newctx, ts)

	client := github.NewClient(tc)

	var commitFiles []github.CommitFile
	if b.Before == "" || b.Before == "0000000000000000000000000000000000000000" {
		response, _, err := client.Repositories.GetCommit(newctx, r.Namespace, r.Name, b.After)
		if err != nil {
			return nil, err
		}
		commitFiles = response.Files
	} else {
		response, _, err := client.Repositories.CompareCommits(newctx, r.Namespace, r.Name, b.Before, b.After)
		if err != nil {
			return nil, err
		}
		commitFiles = response.Files
	}

	var files []string
	for _, f := range commitFiles {
		files = append(files, *f.Filename)
	}
	fmt.Println("github saw these files changed", files)
	return files, nil
}
