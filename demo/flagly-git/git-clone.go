package main

import "github.com/chzyer/flagly"

type GitClone struct {
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
	println("git clone handler")
	return nil
}

func (g *GitClone) FlaglyDesc() string {
	return "Create an empty Git repository or reinitialize an existing one"
}
