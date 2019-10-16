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

type conditions struct {
	Changeset droneyaml.Condition `json:"changeset,omitempty"`
}

type container struct {
	Name string     `json:"name,omitempty"`
	When conditions `json:"when,omitempty"`
}

type pipeline struct {
	Version string `json:"version,omitempty"`
	Kind    string `json:"kind,omitempty"`
	Type    string `json:"type,omitempty"`
	Name    string `json:"name,omitempty"`

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

	manifest, err := droneyaml.ParseRawString(config)
	if err != nil {
		return nil, nil
	}

	cManifest := &droneyaml.Manifest{}

	for _, n := range manifest {
		var dResource droneyaml.Resource
		var cResource droneyaml.Resource
		switch n.Kind {
		case "pipeline":
			dResource = new(droneyaml.Pipeline)
			cResource = new(pipeline)
			err := yaml.Unmarshal(n.Data, dResource)
			if err != nil {
				return nil, nil
			}
			err = yaml.Unmarshal(n.Data, cResource)
			if err != nil {
				return nil, nil
			}

			switch t := cResource.(type) {
			case *pipeline:
				if len(t.Trigger.Changeset.Include) > 0 {
					pipeline := &droneyaml.Pipeline{}
					pipeline.Name = t.Name
					pipeline.Kind = t.Kind
					for _, include := range t.Trigger.Changeset.Include {
						fmt.Println("include is", include)
					}
					if len(t.Steps) > 0 {
						//steps := &droneyaml.Conditions{}
						for _, step := range t.Steps {
							if step == nil {
								continue
							}
							fmt.Println("step is", step.Name)
						}
					}
					cManifest.Resources = append(cManifest.Resources, pipeline)
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
	buf := new(bytes.Buffer)
	pretty.Print(buf, cManifest)
	return &drone.Config{
		Data: buf.String(),
	}, nil
}
