// Copyright 2019 the Drone Authors. All rights reserved.
// Use of this source code is governed by the Blue Oak Model License
// that can be found in the LICENSE file.

package plugin

import (
	"bytes"
	"context"
	"io"

	"github.com/drone/drone-go/drone"
	"github.com/drone/drone-go/plugin/converter"

	"github.com/buildkite/yaml"
	"github.com/sirupsen/logrus"
)

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
		Event condition              `yaml:"event,omitempty"`
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

// New returns a new conversion plugin.
func New(token string) converter.Plugin {
	return &plugin{
		token: token,
	}
}

func (p *plugin) Convert(ctx context.Context, req *converter.Request) (*drone.Config, error) {

	logrus.WithFields(logrus.Fields{
		"action":    req.Build.Action,
		"after":     req.Build.After,
		"before":    req.Build.Before,
		"namespace": req.Repo.Namespace,
		"name":      req.Repo.Name,
	}).Infoln("initiated")

	data := req.Config.Data
	resources, pathsSeen, err := parsePipelines(data, req.Build, req.Repo, p.token)
	if err != nil {
		return nil, nil
	}

	var config string
	if pathsSeen {
		logrus.WithFields(logrus.Fields{
			"action":    req.Build.Action,
			"after":     req.Build.After,
			"before":    req.Build.Before,
			"namespace": req.Repo.Namespace,
			"name":      req.Repo.Name,
		}).Infoln("paths fields were seen, marshaling config")

		c, err := marshal(resources)
		if err != nil {
			return nil, nil
		}
		config = string(c)
	} else {
		logrus.WithFields(logrus.Fields{
			"action":    req.Build.Action,
			"after":     req.Build.After,
			"before":    req.Build.Before,
			"namespace": req.Repo.Namespace,
			"name":      req.Repo.Name,
		}).Infoln("no paths fields seen, no marshaling necessary")

		config = data
	}

	return &drone.Config{
		Data: config,
	}, nil

}
