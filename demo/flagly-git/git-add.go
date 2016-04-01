package main

import "github.com/chzyer/flagly"

type GitAdd struct {
	Git         *Git `flagly:"parent"`
	Update      bool `name:"u" desc:"update tracked files"`
	Interactive bool `name:"i" desc:"interactive picking"`

	PathSpec []string `type:"[]" name:"pathspec"`
}

func (a *GitAdd) FlaglyDesc() string {
	return "Add file contents to the index"
}

func (a *GitAdd) FlaglyHandle() error {
	println("git add")
	return flagly.ErrShowUsage
}
