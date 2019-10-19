package plugin

import (
	"fmt"

	"github.com/drone/drone-go/drone"

	filepath "github.com/bmatcuk/doublestar"
)

// match returns true if the string matches the include
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

// includes returns true if the string matches the include
// patterns.
func (c *condition) includes(v string) bool {
	for _, pattern := range c.Include {
		if ok, _ := filepath.Match(pattern, v); ok {
			return true
		}
	}
	return false
}

// excludes returns true if the string matches the exclude
// patterns.
func (c *condition) excludes(v string) bool {
	for _, pattern := range c.Exclude {
		if ok, _ := filepath.Match(pattern, v); ok {
			return true
		}
	}
	return false
}

func parsePipelines(data string, build drone.Build, repo drone.Repo, token string) ([]*resource, bool, error) {

	resources, err := unmarshal([]byte(data))
	if err != nil {
		return nil, false, err
	}

	pathsSeen := false
	checkedGithub := false
	var changedFiles []string
	for _, resource := range resources {
		switch resource.Kind {
		case "pipeline":
			// there must be a better way to check whether paths.include or paths.exclude is set
			if len(append(resource.Trigger.Paths.Include, resource.Trigger.Paths.Exclude...)) > 0 {
				pathsSeen = true
				if !checkedGithub {
					changedFiles, err = getFilesChanged(repo, build, token)
					if err != nil {
						return nil, false, err
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
					resource.Trigger.Event.Exclude = []string{"*"}
				}
			}

			for _, step := range resource.Steps {
				if step == nil {
					continue
				}
				// there must be a better way to check whether paths.include or paths.exclude is set
				if len(append(step.When.Paths.Include, step.When.Paths.Exclude...)) > 0 {
					pathsSeen = true
					if !checkedGithub {
						changedFiles, err = getFilesChanged(repo, build, token)
						if err != nil {
							return nil, false, err
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
						step.When.Event.Exclude = []string{"*"}
					}
				}
			}
		}
	}
	return resources, pathsSeen, nil
}
