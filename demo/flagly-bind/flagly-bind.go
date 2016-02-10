package main

import (
	"fmt"
	"strconv"

	"github.com/chzyer/flagly"
)

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
	fmt.Printf("config: -v=%v name=%v\n", cfg.Verbose, strconv.Quote(cfg.Name))
}
