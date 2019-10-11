// Copyright 2019 the Drone Authors. All rights reserved.
// Use of this source code is governed by the Blue Oak Model License
// that can be found in the LICENSE file.

package plugin

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"strings"

	bkyaml "github.com/buildkite/yaml"
	"github.com/drone/drone-go/drone"
	"github.com/drone/drone-go/plugin/converter"

	"github.com/drone/drone-yaml/yaml"
	//"github.com/google/go-github/github"
	//"golang.org/x/oauth2"
)

type csConditions struct {
	Changeset yaml.Condition `json:"changeset,omitempty"`
}

type csContainer struct {
	When csConditions `json:"when,omitempty"`
}

type csPipeline struct {
	Version string            `json:"version,omitempty"`
	Kind    string            `json:"kind,omitempty"`
	Type    string            `json:"type,omitempty"`
	Name    string            `json:"name,omitempty"`
	Steps   []*yaml.Container `json:"steps,omitempty"`
	Trigger csConditions      `json:"trigger,omitempty"`
}

func csParse(r io.Reader) (*yaml.Manifest, error) {
	resources, err := yaml.ParseRaw(r)
	if err != nil {
		return nil, err
	}
	manifest := new(yaml.Manifest)
	for _, raw := range resources {
		if raw == nil {
			continue
		}
		resource, err := parseRaw(raw)
		if err != nil {
			return nil, err
		}
		if resource.GetKind() == "" {
			return nil, errors.New("yaml: missing kind attribute")
		}
		manifest.Resources = append(
			manifest.Resources,
			resource,
		)
	}
	return manifest, nil
}

func parseRaw(r *yaml.RawResource) (yaml.Resource, error) {
	var obj yaml.Resource
	switch r.Kind {
	case "pipeline":
		obj = new(csPipeline)
	}
	err := bkyaml.Unmarshal(r.Data, obj)
	return obj, err
}

func csParseBytes(b []byte) (*yaml.Manifest, error) {
	return csParse(
		bytes.NewBuffer(b),
	)
}

func csParseString(s string) (*yaml.Manifest, error) {
	return csParseBytes(
		[]byte(s),
	)
}

func (p *csPipeline) GetVersion() string { return p.Version }

func (p *csPipeline) GetKind() string { return p.Kind }

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
	manifest, err := csParseString(config)
	if err != nil {
		return nil, err
	}
	//var pipeline *csPipeline
	for _, resource := range manifest.Resources {
		v, ok := resource.(*csPipeline)
		if !ok {
			log.Println("notok")
			continue
		}
		fmt.Println(v.Kind)
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
