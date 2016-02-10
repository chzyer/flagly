# flagly

[![Software License](https://img.shields.io/badge/license-MIT-brightgreen.svg)](LICENSE.md)
[![Build Status](https://travis-ci.org/chzyer/flagly.svg?branch=master)](https://travis-ci.org/chzyer/flagly)
[![GoDoc](https://godoc.org/github.com/chzyer/flagly?status.svg)](https://godoc.org/github.com/chzyer/flagly)
[![Gitter](https://badges.gitter.im/Join%20Chat.svg)](https://gitter.im/chzyer/flagly?utm_source=badge&utm_medium=badge&utm_campaign=pr-badge)

The easier way to parsing command-line flag in Golang.

# Usage

```
go get github.com/chzyer/flagly
```

## binding

```{go}
type Config struct {
	Verbose bool   `v desc:"be more verbose"`
	Name    string "[0]"
}

func NewConfig() *Config {
	var cfg Config
	flagly.Bind(&cfg)
	return &cfg
}

func main() {
	cfg := NewConfig()
	fmt.Printf("config: %+v\n", cfg)
}
```

source file: [flagly-binding](https://github.com/chzyer/flagly/blob/master/demo/flagly-binding/flagly-binding.go)

```
$ go install github.com/chzyer/flagly/demo/flagly-binding
$ flagly-binding -v name
config: &{Verbose:true Name:name}
```

## routing

```{go}
type Git struct {
	Version bool `v desc:"show version"`
	
	// sub handlers
	Clone *GitClone `flaglyHandler`
	Init  *GitInit  `flaglyHandler`
}

type GitClone struct {
	Parent *Git `flaglyParent`

	Verbose  bool   `v desc:"be more verbose"`
	Quiet    bool   `q desc:"be more quiet"`
	Progress bool   `progress desc:"force progress reporting"`
	Template string `arg:"template-directory"`

	Repo string `[0]`
	Dir  string `[1] default`
}

func (g *GitClone) FlaglyHandle() error {
	if g.Repo == "" {
		return flagly.ErrShowUsage
	}
	fmt.Printf("git clone %+v %+v\n", g.Parent, g)
	return nil
}

func (g *GitClone) FlaglyDesc() string {
	return "Create an empty Git repository or reinitialize an existing one"
}

type GitInit struct {
	Quiet bool `q desc:"be quiet"`
}

func (g *GitInit) FlaglyDesc() string {
	return "Clone a repository into a new directory"
}

func main() {
	var git Git
	flagly.Run(&git)
}
```

source file: [flagly-git](https://github.com/chzyer/flagly/blob/master/demo/flagly-git/flagly-git.go)

```
$ go install github.com/chzyer/flagly/demo/flagly-git
$ flagly-git
usage: flagly-git [option] <command>

options:
    -v                  show version
    -h                  show help

commands:
    clone               Create an empty Git repository or reinitialize an existing one
    init                Clone a repository into a new directory
	
$ flagly-get -v clone -h

usage: flagly-git [flagly-git option] clone [option] [--] <repo> [<dir>]

options:
    -v                  be more verbose
    -q                  be more quiet
    -progress           force progress reporting
    -template <template-directory>
                        directory from which templates will be used
    -h                  show help

flagly-git options:
    -v                  show version
    -h                  show help
$ flagly-git -v clone repoName
git clone
    &{Version:true Clone:<nil> Init:<nil> Add:<nil>}
    &{Parent:0xc20801e220 Verbose:false Quiet:false Progress:false Template: Repo:repoName Dir:}
```
