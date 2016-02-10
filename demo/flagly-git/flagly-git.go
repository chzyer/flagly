package main

import (
	"os"

	"github.com/chzyer/flagly"
)

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
	obj, err := flagly.Compile(os.Args[0], &git)
	if err != nil {
		println(err.Error())
		os.Exit(1)
	}
	if err := obj.Run(os.Args[1:]); err != nil {
		println(err.Error())
	}
}
