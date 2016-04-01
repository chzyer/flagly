# flagly

[![Software License](https://img.shields.io/badge/license-MIT-brightgreen.svg)](LICENSE.md)
[![Build Status](https://travis-ci.org/chzyer/flagly.svg?branch=master)](https://travis-ci.org/chzyer/flagly)
[![GoDoc](https://godoc.org/github.com/chzyer/flagly?status.svg)](https://godoc.org/github.com/chzyer/flagly)
[![codebeat badge](https://codebeat.co/badges/4efbd40b-4d84-48f5-8363-df06c2e9b241)](https://codebeat.co/projects/github-com-chzyer-flagly)

The easier way to parsing command-line flag in Golang, also building a command line app.

It can also provides shell-like interactives by using [readline](https://github.com/chzyer/readline) (demo: [flagly-shell](https://github.com/chzyer/flagly/blob/master/demo/flagly-shell/flagly-shell.go))

# Usage

```
go get github.com/chzyer/flagly
```

## binding

```{go}
type Config struct {
	Verbose bool   `name:"v" desc:"be more verbose"`
	Name    string `type:"[0]"`
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
	Version bool `name:"v" desc:"show version"`
	
	// sub handlers
	Clone *GitClone `flagly:"handler"`
	Init  *GitInit  `flagly:"handler"`
}

type GitClone struct {
	Parent *Git `flagly:"parent"`

	Verbose  bool   `name:"v" desc:"be more verbose"`
	Quiet    bool   `name:"q" desc:"be more quiet"`
	Progress bool   `name:"progress" desc:"force progress reporting"`
	Template string `arg:"template-directory"`

	Repo string `name:"[0]"`
	Dir  string `name:"[1]" default:""`
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
	Quiet bool `name:"q" desc:"be quiet"`
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

```{shell}
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

# Feedback

If you have any questions, please submit a github issue and any pull requests is welcomed :)

* [https://twitter.com/chzyer](https://twitter.com/chzyer)  
* [http://weibo.com/2145262190](http://weibo.com/2145262190)  
