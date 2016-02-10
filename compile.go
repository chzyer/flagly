package flagly

import "errors"

var (
	ErrMustAPtrToStruct = errors.New("must a pointer to struct")
	ErrMustAStruct      = errors.New("must a struct")
)
