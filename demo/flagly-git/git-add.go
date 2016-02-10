package main

import "github.com/chzyer/flagly"

type GitAdd struct {
	Git         *Git "flaglyParent"
	Update      bool `u desc:"update tracked files"`
	Interactive bool `i desc:"interactive picking"`

	PathSpec []string `[] arg:"pathspec"`
}

func (a *GitAdd) FlaglyDesc() string {
	return "Add file contents to the index"
}

func (a *GitAdd) FlaglyHandle() error {
	println("git add")
	return flagly.ErrShowUsage
}
