// Copyright 2019 the Drone Authors. All rights reserved.
// Use of this source code is governed by the Blue Oak Model License
// that can be found in the LICENSE file.

package plugin

import (
	"context"
	"fmt"
	"strings"

	bkyaml "github.com/buildkite/yaml"
	"github.com/drone/drone-go/drone"
	"github.com/drone/drone-go/plugin/converter"
	yaml "github.com/drone/drone-yaml/yaml"
	//"github.com/google/go-github/github"
	//"golang.org/x/oauth2"
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

type conditions struct {
	Changeset yaml.Condition `json:"changeset,omitempty"`
}

type container struct {
	Name string     `json:"name,omitempty"`
	When conditions `json:"when,omitempty"`
}

type pipeline struct {
	Version string       `json:"version,omitempty"`
	Kind    string       `json:"kind,omitempty"`
	Type    string       `json:"type,omitempty"`
	Steps   []*container `json:"steps,omitempty"`
	Trigger conditions   `json:"trigger,omitempty"`
}

// GetVersion returns the resource version.
func (p *pipeline) GetVersion() string { return p.Version }

// GetKind returns the resource kind.
func (p *pipeline) GetKind() string { return p.Kind }

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

	manifest, err := yaml.ParseRawString(config)
	if err != nil {
		return nil, nil
	}

	for _, n := range manifest {
		var obj yaml.Resource
		switch n.Kind {
		case "pipeline":
			obj = new(pipeline)
			err := bkyaml.Unmarshal(n.Data, obj)
			if err != nil {
				return nil, nil
			}
			switch t := obj.(type) {
			case *pipeline:
				if len(t.Trigger.Changeset.Include) > 0 {
					fmt.Println("I see trigger include")
				}
				if len(t.Steps) > 0 {
					fmt.Println("there are steps")
					for _, step := range t.Steps {
						if step == nil {
							continue
						}
						fmt.Println(step.Name)
					}
				}
			}
		}
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
	return &drone.Config{
		Data: config,
	}, nil
}
