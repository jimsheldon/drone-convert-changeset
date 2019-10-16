// Copyright 2019 the Drone Authors. All rights reserved.
// Use of this source code is governed by the Blue Oak Model License
// that can be found in the LICENSE file.

package plugin

import (
	"bytes"
	"context"
	"fmt"
	"strings"

	"github.com/drone/drone-go/drone"
	"github.com/drone/drone-go/plugin/converter"
	droneyaml "github.com/drone/drone-yaml/yaml"
	"github.com/drone/drone-yaml/yaml/pretty"

	"github.com/buildkite/yaml"
	//"github.com/google/go-github/github"
	//"golang.org/x/oauth2"
)

const (
	separator = "---"
	newline   = "\n"
)

// New returns a new conversion plugin.
func New(token string) converter.Plugin {
	return &plugin{
		token: token,
	}
}

type plugin struct {
	token string
}

func (p *plugin) Convert(ctx context.Context, req *converter.Request) (*drone.Config, error) {
	// TODO this should be modified or removed. For
	// demonstration purposes we show how you can ignore
	// certain configuration files by file extension.
	if !strings.HasSuffix(req.Repo.Config, ".yml") {
		// a nil response instructs the Drone server to
		// use the configuration file as-is, without
		// modification.
		return nil, nil
	}

	// get the configuration file from the request.
	config := req.Config.Data

	manifest, err := droneyaml.ParseRawString(config)
	if err != nil {
		return nil, nil
	}

	newManifest := &droneyaml.Manifest{}

	for _, n := range manifest {
		var r droneyaml.Resource
		switch n.Kind {
		case "cron":
			r = new(droneyaml.Cron)
		case "secret":
			r = new(droneyaml.Secret)
		case "signature":
			r = new(droneyaml.Signature)
		case "registry":
			r = new(droneyaml.Registry)
		default:
			r = new(droneyaml.Pipeline)
			err := yaml.Unmarshal(n.Data, r)
			if err != nil {
				return nil, nil
			}

			switch t := r.(type) {
			case *droneyaml.Pipeline:
				if len(t.Trigger.Paths.Include) > 0 {
					for _, include := range t.Trigger.Paths.Include {
						fmt.Println("include is", include)
					}
					t.Trigger.Event.Exclude = []string{"*"}
					if len(t.Steps) > 0 {
						for _, step := range t.Steps {
							if step == nil {
								continue
							}
							fmt.Println("step is", step.Name)
						}
					}
				}
			}
		}
		newManifest.Resources = append(newManifest.Resources, r)
	}

	// TODO this should be modified or removed. For
	// demonstration purposes we make a simple modification
	// to the configuration file and add a newline.
	//config = config + "\nwoot"

	//newctx := context.Background()
	//ts := oauth2.StaticTokenSource(
	//	&oauth2.Token{AccessToken: p.token},
	//)
	//tc := oauth2.NewClient(newctx, ts)
	//client := github.NewClient(tc)
	//repos, _, _ := client.Repositories.List(newctx, "", nil)
	//log.Println(repos)

	// returns the modified configuration file.
	buf := new(bytes.Buffer)
	pretty.Print(buf, newManifest)
	return &drone.Config{
		Data: buf.String(),
	}, nil
}
