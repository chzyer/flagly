package main

import (
	"encoding/base64"
	"errors"
	"os"
	"time"

	"github.com/chzyer/flagly"
	"github.com/chzyer/readline"
)

type Help struct{}

func (Help) FlaglyHandle(h *flagly.Handler) error {
	return errors.New(h.Parent.Usage(""))
}

// -----------------------------------------------------------------------------

type Time struct {
	Layout string `[0] default:"2006-01-02 15:04:05"`
}

func (t *Time) FlaglyHandle() error {
	now := time.Now()
	println(now.Format(t.Layout))
	return nil
}

type Base64 struct {
	IsDecode bool   `d desc:"decode string"`
	String   string `[0]`
}

func (b *Base64) FlaglyHandle() error {
	if b.IsDecode {
		ret, err := base64.URLEncoding.DecodeString(b.String)
		if err != nil {
			return err
		}
		println(string(ret))
	} else {
		ret := base64.URLEncoding.EncodeToString([]byte(b.String))
		println(ret)
	}
	return nil
}

// -----------------------------------------------------------------------------

type Program struct {
	Help   *Help   `flaglyHandler`
	Time   *Time   `flaglyHandler`
	Base64 *Base64 `flaglyHandler`
}

func main() {
	rl, err := readline.NewEx(&readline.Config{
		Prompt:      "> ",
		HistoryFile: "/tmp/flagly-shell.readline",
	})
	if err != nil {
		println(err.Error())
		os.Exit(1)
	}
	defer rl.Close()

	var p Program
	fset, err := flagly.Compile("", &p)
	if err != nil {
		println(err.Error())
		os.Exit(1)
	}

	for {
		line, err := rl.Readline()
		if err != nil {
			break
		}
		if line == "" {
			continue
		}
		if err := fset.Run(flagly.SplitArgs(line)); err != nil {
			println(err.Error())
		}
	}
}
