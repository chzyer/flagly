package main

import (
	"fmt"

	"github.com/chzyer/flagly"
)

type GitClone struct {
	Parent *Git `flaglyParent`

	Verbose  bool   `v desc:"be more verbose"`
	Quiet    bool   `q desc:"be more quiet"`
	Progress bool   `progress desc:"force progress reporting"`
	Template string `arg:"template-directory"`

	Repo string `[0]`
	Dir  string `[1] default`
}

func (g *GitClone) FlaglyInit() {
	flagly.SetDesc(&g.Template, "directory from which templates will be used")
}

func (g *GitClone) FlaglyHandle() error {
	if g.Repo == "" {
		return flagly.Error("error: repo is empty")
	}

	fmt.Printf("git clone\n    %+v\n    %+v\n", g.Parent, g)
	return nil
}

func (g *GitClone) FlaglyDesc() string {
	return "Create an empty Git repository or reinitialize an existing one"
}
