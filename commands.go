package flagly

import (
	"errors"
)

type CmdHelp struct{}

func NewHandlerHelp() *Handler {
	h := NewHandler("help")
	h.CompileIface(&CmdHelp{})
	return h
}

func (CmdHelp) FlaglyHandle(h *Handler) error {
	return errors.New(h.GetRoot().Usage(""))
}

func (CmdHelp) FlaglyDesc() string {
	return "show help"
}
