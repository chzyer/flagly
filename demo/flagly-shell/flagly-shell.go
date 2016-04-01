/*
usage:
    $ go install github.com/chzyer/flagly/demo/flagly-shell
    $ flagly-shell
    > help
    commands:
        help
        time
        base64
    > time
    2016-02-11 12:12:43
    > time -h
    usage: time [option] [--] [<layout>]

    options:
        -h                  show help
    > base64
    missing content

    usage: base64 [option] [--] <content>

    options:
        -d                  decode string
        -h                  show help
    > base64 hello
    aGVsbG8=
*/
package main

import (
	"encoding/base64"
	"errors"
	"os"
	"time"

	"github.com/chzyer/flagly"
	"github.com/chzyer/readline"
	"github.com/google/shlex"
)

type Help struct{}

func (Help) FlaglyHandle(h *flagly.Handler) error {
	return errors.New(h.Parent.Usage(""))
}

// -----------------------------------------------------------------------------

type Time struct {
	Layout string `type:"[0]" default:"2006-01-02 15:04:05"`
}

func (t *Time) FlaglyHandle() error {
	now := time.Now()
	println(now.Format(t.Layout))
	return nil
}

type Base64 struct {
	IsDecode bool   `name:"d" desc:"decode string"`
	Content  string `type:"[0]"`
}

func (b *Base64) FlaglyHandle() error {
	if b.Content == "" {
		return flagly.Error("missing content")
	}
	if b.IsDecode {
		ret, err := base64.URLEncoding.DecodeString(b.Content)
		if err != nil {
			return err
		}
		println(string(ret))
	} else {
		ret := base64.URLEncoding.EncodeToString([]byte(b.Content))
		println(ret)
	}
	return nil
}

// -----------------------------------------------------------------------------

type Program struct {
	Help   *Help   `flagly:"handler"`
	Time   *Time   `flagly:"handler"`
	Base64 *Base64 `flagly:"handler"`
	Level1 *Level1 `flagly:"handler"`
}

type Level1 struct {
	Level2 *Level2 `flagly:"handler"`
}

type Level2 struct {
	Level3 *Level3 `flagly:"handler"`
}

type Level3 struct{}

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
	rl.Config.AutoComplete = &readline.SegmentComplete{fset.Completer()}

	for {
		line, err := rl.Readline()
		if err != nil {
			break
		}
		if line == "" {
			continue
		}
		command, err := shlex.Split(line)
		if err != nil {
			println("error: " + err.Error())
			continue
		}
		if err := fset.Run(command); err != nil {
			println(err.Error())
		}
	}
}
