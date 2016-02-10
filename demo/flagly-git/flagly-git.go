package main

import "github.com/chzyer/flagly"

type Git struct {
	Version bool "v"

	Clone *GitClone "flaglyHandler"
	Init  *GitInit  "flaglyHandler"

	Add *GitAdd "flaglyHandler"
}

func (g *Git) FlaglyInit() {
	flagly.SetDesc(&g.Version, "show version")
}

func main() {
	var git Git
	flagly.Run(&git)
}
