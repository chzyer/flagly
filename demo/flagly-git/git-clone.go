package main

import (
	"fmt"

	"github.com/chzyer/flagly"
)

type GitClone struct {
	// we can get struct `Git` directly by presenting `flaglyParent`,
	Parent *Git `flagly:"parent"`

	Verbose  bool   `name:"v" desc:"be more verbose"`
	Quiet    bool   `name:"q" desc:"be more quiet"`
	Progress bool   `name:"progress" desc:"force progress reporting"`
	Template string `arg:"template-directory"`

	Repo string `type:"[0]"`
	Dir  string `type:"[1]" default:"."`
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
