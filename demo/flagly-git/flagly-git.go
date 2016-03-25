package main

import "github.com/chzyer/flagly"

type Git struct {
	// flags for flagly-git
	Version bool `name:"v"`

	// sub handlers must be specified `flaglyHandler` tag
	Clone *GitClone `flagly:"handler"`
	Init  *GitInit  `flagly:"handler"`

	Add *GitAdd `flagly:"handler"`
}

func (g *Git) FlaglyInit() {
	// we can set the description via `flagly.SetDesc` or just a field tag `desc:"xxx"`
	flagly.SetDesc(&g.Version, "show version")
}

func main() {
	var git Git
	flagly.Run(&git)
}
