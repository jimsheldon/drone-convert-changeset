// Copyright 2019 the Drone Authors. All rights reserved.
// Use of this source code is governed by the Blue Oak Model License
// that can be found in the LICENSE file.

package plugin

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/drone/drone-go/drone"
	"github.com/drone/drone-go/plugin/converter"
	yaml "github.com/drone/drone-yaml/yaml"
	goyaml "gopkg.in/yaml.v2"
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
	yaml.Conditions
	Changeset yaml.Condition `json:"changeset,omitempty"`
}

type container struct {
	Build       *yaml.Build                `json:"build,omitempty"`
	Command     []string                   `json:"command,omitempty"`
	Commands    []string                   `json:"commands,omitempty"`
	Detach      bool                       `json:"detach,omitempty"`
	DependsOn   []string                   `json:"depends_on,omitempty" yaml:"depends_on"`
	Devices     []*yaml.VolumeDevice       `json:"devices,omitempty"`
	DNS         []string                   `json:"dns,omitempty"`
	DNSSearch   []string                   `json:"dns_search,omitempty" yaml:"dns_search"`
	Entrypoint  []string                   `json:"entrypoint,omitempty"`
	Environment map[string]*yaml.Variable  `json:"environment,omitempty"`
	ExtraHosts  []string                   `json:"extra_hosts,omitempty" yaml:"extra_hosts"`
	Failure     string                     `json:"failure,omitempty"`
	Image       string                     `json:"image,omitempty"`
	Network     string                     `json:"network_mode,omitempty" yaml:"network_mode"`
	Name        string                     `json:"name,omitempty"`
	Ports       []*yaml.Port               `json:"ports,omitempty"`
	Privileged  bool                       `json:"privileged,omitempty"`
	Pull        string                     `json:"pull,omitempty"`
	Push        *yaml.Push                 `json:"push,omitempty"`
	Resources   *yaml.Resources            `json:"resources,omitempty"`
	Settings    map[string]*yaml.Parameter `json:"settings,omitempty"`
	Shell       string                     `json:"shell,omitempty"`
	User        string                     `json:"user,omitempty"`
	Volumes     []*yaml.VolumeMount        `json:"volumes,omitempty"`
	When        conditions                 `json:"when,omitempty"`
	WorkingDir  string                     `json:"working_dir,omitempty" yaml:"working_dir"`
}

type pipeline struct {
	Version string `json:"version,omitempty"`
	Kind    string `json:"kind,omitempty"`
	Type    string `json:"type,omitempty"`
	Name    string `json:"name,omitempty"`

	Clone       yaml.Clone        `json:"clone,omitempty"`
	Concurrency yaml.Concurrency  `json:"concurrency,omitempty"`
	DependsOn   []string          `json:"depends_on,omitempty" yaml:"depends_on" `
	Node        map[string]string `json:"node,omitempty" yaml:"node"`
	Platform    yaml.Platform     `json:"platform,omitempty"`
	PullSecrets []string          `json:"image_pull_secrets,omitempty" yaml:"image_pull_secrets"`
	Services    []*yaml.Container `json:"services,omitempty"`
	Steps       []*container      `json:"steps,omitempty"`
	Trigger     conditions        `json:"trigger,omitempty"`
	Volumes     []*yaml.Volume    `json:"volumes,omitempty"`
	Workspace   yaml.Workspace    `json:"workspace,omitempty"`
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

	var pipeline pipeline

	// get the configuration file from the request.
	config := req.Config.Data

	//var b StructB

	err := goyaml.Unmarshal([]byte(config), &pipeline)
	if err != nil {
		log.Fatalf("cannot unmarshal data: %v", err)
	}

	resources, err := yaml.ParseRawString(config)
	if err != nil {
		return nil, nil
	}
	for _, raw := range resources {
		fmt.Println(raw.Kind)
	}
	//fmt.Printf("%v", pipeline.Trigger.Changeset.Include)
	//for _, step := range pipeline.Steps {
	//	fmt.Println(step.Image)
	//}

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
