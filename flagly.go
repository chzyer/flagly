package flagly

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"reflect"
	"sync"
)

const (
	handlerPkgPath = "flagly.Handler"
	flaglyHandler  = "flaglyHandler"
)

var (
	ErrMustAFunc         = errors.New("argument must be a func")
	ErrFuncOutMustAError = errors.New("func.out must be a error")

	ErrMustAPtrToStruct = errors.New("must a pointer to struct")
	ErrMustAStruct      = errors.New("must a struct")
)

func Exit(info interface{}) {
	println(fmt.Sprint(info))
	os.Exit(2)
}

func Bind(target interface{}) {
	if err := BindByArgs(target, os.Args); err != nil {
		Exit(err)
	}
}

func Run(target interface{}) {
	if err := RunByArgs(target, os.Args); err != nil {
		Exit(err)
	}
}

func BindByArgs(target interface{}, args []string) error {
	fset, err := Compile(args[0], target)
	if err != nil {
		return err
	}
	ptr := reflect.ValueOf(target)
	if err := fset.Bind(ptr, args[1:]); err != nil {
		return err
	}
	return nil
}

func RunByArgs(target interface{}, args []string) error {
	fset, err := Compile(args[0], target)
	if err != nil {
		return err
	}
	if err := fset.Run(args[1:]); err != nil {
		return err
	}
	return nil
}

func Compile(name string, target interface{}) (*FlaglySet, error) {
	fset := New(name)
	if err := fset.Compile(target); err != nil {
		return nil, err
	}
	return fset, nil
}

type FlaglySet struct {
	subHandler *Handler
}

func New(name string) *FlaglySet {
	fset := &FlaglySet{
		subHandler: NewHandler(name),
	}
	return fset
}

func (f *FlaglySet) Compile(target interface{}) error {
	return f.subHandler.Compile(reflect.TypeOf(target))
}

func (f *FlaglySet) Add(h *Handler) {
	f.subHandler.AddHandler(h)
	return
}

func (f *FlaglySet) SetHandleFunc(hf interface{}) error {
	return f.subHandler.SetHandleFunc(hf)
}

func (f *FlaglySet) Bind(value reflect.Value, args []string) error {
	return f.subHandler.Bind(value, args)
}

func (f *FlaglySet) Run(args []string) (err error) {
	stack := []reflect.Value{}
	if err = f.subHandler.Run(&stack, args); err != nil {
		return err
	}
	return
}

func (f *FlaglySet) GetHandler(name string) *Handler {
	return f.subHandler.GetHandler(name)
}

func (f *FlaglySet) Usage() string {
	buffer := bytes.NewBuffer(nil)
	f.subHandler.usage(buffer, "")
	return buffer.String()
}

func IsShowUsage(err error) *showUsageError {
	if s, ok := err.(*showUsageError); ok {
		return s
	}
	if s, ok := err.(showUsageError); ok {
		return &s
	}
	return nil
}

// -----------------------------------------------------------------------------

var (
	descMap   = map[uintptr]string{}
	descGuard sync.Mutex
)

func SetDesc(target interface{}, desc string) {
	value := reflect.ValueOf(target)
	if value.Kind() != reflect.Ptr {
		panic("SetDesc only accept pointer")
	}
	value = reflect.ValueOf(target).Elem()
	descGuard.Lock()
	descMap[value.UnsafeAddr()] = desc
	descGuard.Unlock()
}

func getDescMap() map[uintptr]string {
	descGuard.Lock()
	ret := descMap
	descMap = map[uintptr]string{}
	descGuard.Unlock()
	return ret
}

func SplitArgs(content string) []string {
	r := bytes.NewReader([]byte(content))
	var EmptyQuote = rune(0)
	var ret []string

	var quote rune
	// 0: not quote char, 1: quote start, 2: quote end
	isQuote := func(ch rune) int {
		if quote == EmptyQuote {
			if ch == '"' || ch == '\'' {
				return 1
			}
			return 0
		}
		if ch == quote {
			return 2
		}
		return 0
	}
	isInQuote := func(ch rune) bool {
		return quote != EmptyQuote
	}
	isSpace := func(ch rune) bool {
		return ch == ' '
	}

	var buf []rune
	for {
		ch, _, err := r.ReadRune()
		if err != nil {
			break
		}
		chType := isQuote(ch)
		if chType == 0 {
			if isSpace(ch) && !isInQuote(ch) {
				if len(buf) > 0 {
					ret = append(ret, string(buf))
					buf = buf[:0]
				}
			} else {
				buf = append(buf, ch)
			}
		} else if chType == 2 { // append even meet empty string
			ret = append(ret, string(buf))
			buf = buf[:0]
			quote = EmptyQuote
		} else { // chType == 1
			quote = ch
		}
	}
	if len(buf) > 0 {
		ret = append(ret, string(buf))
	}
	return ret
}
