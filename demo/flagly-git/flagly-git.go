package main

import "github.com/chzyer/flagly"

type Git struct {
	// flags for flagly-git
	Version bool "v"

	// sub handlers must be specified `flaglyHandler` tag
	Clone *GitClone "flaglyHandler"
	Init  *GitInit  "flaglyHandler"

	Add *GitAdd "flaglyHandler"
}

func (g *Git) FlaglyInit() {
	// we can set the description via `flagly.SetDesc` or just a field tag `desc:"xxx"`
	flagly.SetDesc(&g.Version, "show version")
}

func main() {
	var git Git
	flagly.Run(&git)
}
