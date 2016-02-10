package flagly

import (
	"bytes"
	"errors"
	"reflect"
	"sync"
)

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

func (f *FlaglySet) Run(args []string) (err error) {
	stack := []reflect.Value{}
	if err = f.subHandler.Run(&stack, args); err != nil {
		if e := IsShowUsage(err); e != nil {
			err = errors.New(e.Usage())
		}
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
