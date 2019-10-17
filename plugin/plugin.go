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
	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
	"gopkg.in/yaml.v2"
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
		Paths condition              `yaml:"paths,omitempty"`
		Attrs map[string]interface{} `yaml:",inline"`
	}

	condition struct {
		Exclude []string `yaml:"exclude,omitempty"`
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

// Match returns true if the string matches the include
// patterns and does not match any of the exclude patterns.
func (c *condition) match(v string) bool {
	if c.excludes(v) {
		return false
	}
	if c.includes(v) {
		return true
	}
	if len(c.Include) == 0 {
		return true
	}
	return false
}

// Includes returns true if the string matches the include
// patterns.
func (c *condition) includes(v string) bool {
	for _, pattern := range c.Include {
		if ok, _ := filepath.Match(pattern, v); ok {
			return true
		}
	}
	return false
}

// Excludes returns true if the string matches the exclude
// patterns.
func (c *condition) excludes(v string) bool {
	for _, pattern := range c.Exclude {
		if ok, _ := filepath.Match(pattern, v); ok {
			return true
		}
	}
	return false
}

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

func (p *plugin) Convert(ctx context.Context, req *converter.Request) (*drone.Config, error) {
	// get the configuration file from the request.
	config := req.Config.Data

	resources, err := unmarshal([]byte(config))
	if err != nil {
		return nil, nil
	}

	checkedGithub := false
	var changedFiles []string
	for _, resource := range resources {
		switch resource.Kind {
		case "pipeline":
			// there must be a better way to check whether paths.include or paths.exclude is set
			if len(append(resource.Trigger.Paths.Include, resource.Trigger.Paths.Exclude...)) > 0 {
				if !checkedGithub {
					changedFiles, err = getFilesChanged(req.Repo, req.Build, p.token)
					if err != nil {
						return nil, nil
					}
					checkedGithub = true
				}
				skipPipeline := true
				for _, p := range changedFiles {
					got, want := resource.Trigger.Paths.match(p), true
					if got == want {
						fmt.Println("keeping pipeline", resource.Attrs["name"])
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
				// there must be a better way to check whether paths.include or paths.exclude is set
				if len(append(step.When.Paths.Include, step.When.Paths.Exclude...)) > 0 {
					if !checkedGithub {
						changedFiles, err = getFilesChanged(req.Repo, req.Build, p.token)
						if err != nil {
							return nil, nil
						}
						checkedGithub = true
					}
					skipStep := true
					for _, i := range changedFiles {
						got, want := step.When.Paths.match(i), true
						if got == want {
							fmt.Println("keeping step", step.Attrs["name"])
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

	newConfig, err := marshal(resources)
	if err != nil {
		return nil, nil
	}

	return &drone.Config{
		Data: string(newConfig),
	}, nil
}
