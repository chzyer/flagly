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

## struct binding

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

## struct routing

```{go}
type Git struct {
	Version bool `v`
	
	// sub handlers
	Clone *GitClone `flaglyHandler`
	Init  *GitInit `flaglyHandler`
}

type GitClone struct {

}

type GitInit struct {
	Quiet bool `q desc:"be quiet"`
}
```
