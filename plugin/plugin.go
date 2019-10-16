// Copyright 2019 the Drone Authors. All rights reserved.
// Use of this source code is governed by the Blue Oak Model License
// that can be found in the LICENSE file.

package plugin

import (
	"bytes"
	"context"
	"fmt"
	"io"

	"github.com/drone/drone-go/drone"
	"github.com/drone/drone-go/plugin/converter"

	filepath "github.com/bmatcuk/doublestar"
	"gopkg.in/yaml.v2"
	//"github.com/google/go-github/github"
	//"golang.org/x/oauth2"
)

// New returns a new conversion plugin.
func New(token string) converter.Plugin {
	return &plugin{
		token: token,
	}
}

type (
	plugin struct {
		token string
	}

	resource struct {
		Kind    string
		Type    string
		Steps   []*step                `yaml:"steps,omitempty"`
		Trigger conditions             `yaml:"trigger,omitempty"`
		Attrs   map[string]interface{} `yaml:",inline"`
	}

	step struct {
		When  conditions             `yaml:"when,omitempty"`
		Attrs map[string]interface{} `yaml:",inline"`
	}

	conditions struct {
		Paths condition              `json:"yaml,omitempty"`
		Attrs map[string]interface{} `yaml:",inline"`
	}

	condition struct {
		Include []string `yaml:"include,omitempty"`
	}
)

func unmarshal(b []byte) ([]*resource, error) {
	buf := bytes.NewBuffer(b)
	res := []*resource{}
	dec := yaml.NewDecoder(buf)
	for {
		out := new(resource)
		err := dec.Decode(out)
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		res = append(res, out)
	}
	return res, nil
}

func marshal(in []*resource) ([]byte, error) {
	buf := new(bytes.Buffer)
	enc := yaml.NewEncoder(buf)
	for _, res := range in {
		err := enc.Encode(res)
		if err != nil {
			return nil, err
		}
	}
	return buf.Bytes(), nil
}

func includes(c condition, i string) bool {
	for _, pattern := range c.Include {
		if ok, _ := filepath.Match(pattern, i); ok {
			fmt.Println("file ", i, " matches pattern ", pattern)
			return true
		}
	}
	return false
}

func (p *plugin) Convert(ctx context.Context, req *converter.Request) (*drone.Config, error) {
	// get the configuration file from the request.
	config := req.Config.Data

	resources, err := unmarshal([]byte(config))

	if err != nil {
		return nil, nil
	}

	changedFiles := [2]string{"file.txt", "directory/thing.txt"}

	for _, resource := range resources {
		switch resource.Kind {
		case "pipeline":
			if len(resource.Trigger.Paths.Include) > 0 {
				skipPipeline := true
				for _, p := range changedFiles {
					if includes(resource.Trigger.Paths, p) {
						skipPipeline = false
						break
					}
				}
				if skipPipeline {
					resource.Trigger.Attrs["event"] = map[string][]string{"exclude": []string{"*"}}
				}
			}

			for _, step := range resource.Steps {
				if step == nil {
					continue
				}

				if len(step.When.Paths.Include) > 0 {
					skipStep := true
					for _, i := range changedFiles {
						if includes(step.When.Paths, i) {
							skipStep = false
							break
						}
					}
					if skipStep {
						step.Attrs["event"] = map[string][]string{"exclude": []string{"*"}}
					}
				}
			}
		}
	}

	//newctx := context.Background()
	//ts := oauth2.StaticTokenSource(
	//	&oauth2.Token{AccessToken: p.token},
	//)
	//tc := oauth2.NewClient(newctx, ts)
	//client := github.NewClient(tc)
	//repos, _, _ := client.Repositories.List(newctx, "", nil)
	//log.Println(repos)

	// returns the modified configuration file.
	//buf := new(bytes.Buffer)
	//pretty.Print(buf, newManifest)

	newConfig, err := marshal(resources)
	return &drone.Config{
		Data: string(newConfig),
	}, nil
}
