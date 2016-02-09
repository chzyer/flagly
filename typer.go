package flagly

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"
)

var (
	types    = map[string]Typer{}
	nilValue reflect.Value
)

func init() {
	RegisterAll(Bool{}, String{}, Duration{}, Int{})
	Register(MapStringString{})
}

func getTypeName(t reflect.Type) string {
	return t.String()
}

func RegisterAll(objs ...BaseTyperParser) {
	for _, obj := range objs {
		Register(ValueWrap{obj}, SliceWrap{obj})
	}
}

func Register(objs ...Typer) {
	for _, obj := range objs {
		types[getTypeName(obj.Type())] = obj
	}
}

func GetTyper(t reflect.Type) (Typer, error) {
	name := getTypeName(t)
	ret := types[name]
	if ret != nil {
		return ret, nil
	}
	return nil, fmt.Errorf("unknown type: %v", name)
}

type BaseTyper interface {
	Type() reflect.Type
	NumArgs() (min int, max int)
	CanBeValue(arg string) bool
}
type BaseTypeArgNamer interface {
	ArgName() string
}
type BaseTyperParser interface {
	BaseTyper
	ParseArgs(args []string) (reflect.Value, error)
}

type SliceWrap struct {
	BaseTyperParser
}

func (s SliceWrap) Type() reflect.Type {
	return reflect.SliceOf(s.BaseTyperParser.Type())
}
func (s SliceWrap) Set(source reflect.Value, args []string) error {
	if len(args) == 0 {
		return nil
	}
	arg, err := s.BaseTyperParser.ParseArgs(args)
	if err != nil {
		return err
	}
	val := reflect.New(s.BaseTyperParser.Type())
	SetToSource(val, arg)
	source.Set(reflect.Append(source, val.Elem()))
	return nil
}

type Typer interface {
	BaseTyper
	Set(source reflect.Value, args []string) error
}

type ValueWrap struct{ BaseTyperParser }

func SetToSource(source, val reflect.Value) {
	if source.Kind() == reflect.Ptr {
		source = source.Elem()
	}
	source.Set(val)
}

func (b ValueWrap) ArgName() string {
	if a, ok := b.BaseTyperParser.(BaseTypeArgNamer); ok {
		return a.ArgName()
	}
	return ""
}

func (b ValueWrap) Set(source reflect.Value, args []string) error {
	val, err := b.ParseArgs(args)
	if err != nil {
		return err
	}
	SetToSource(source, val)
	return nil
}

type Int struct{}

func (Int) Type() reflect.Type     { return reflect.TypeOf(int(0)) }
func (Int) CanBeValue(string) bool { return true }
func (Int) NumArgs() (int, int)    { return 1, 1 }
func (Int) ArgName() string        { return "number" }
func (Int) ParseArgs(args []string) (reflect.Value, error) {
	val, err := strconv.Atoi(args[0])
	if err != nil {
		return nilValue, err
	}
	return reflect.ValueOf(val), nil
}

type String struct{}

func (String) Type() reflect.Type     { return reflect.TypeOf("") }
func (String) CanBeValue(string) bool { return true }
func (String) NumArgs() (int, int)    { return 1, 1 }
func (String) ParseArgs(args []string) (reflect.Value, error) {
	return reflect.ValueOf(args[0]), nil
}

type Bool struct{}

func (Bool) Type() reflect.Type  { return reflect.TypeOf(true) }
func (Bool) NumArgs() (int, int) { return 0, 1 }
func (Bool) CanBeValue(arg string) bool {
	return arg == "true" || arg == "false"
}
func (Bool) ParseArgs(args []string) (reflect.Value, error) {
	val := true
	if len(args) == 1 && args[0] == "false" {
		val = false
	}
	return reflect.ValueOf(val), nil
}

type Duration struct{}

func (Duration) Type() reflect.Type     { return reflect.TypeOf(time.Second) }
func (Duration) NumArgs() (int, int)    { return 1, 1 }
func (Duration) CanBeValue(string) bool { return true }
func (Duration) ParseArgs(args []string) (reflect.Value, error) {
	a, err := time.ParseDuration(args[0])
	if err != nil {
		return nilValue, err
	}
	return reflect.ValueOf(a), nil
}

type MapStringString struct{}

func (MapStringString) Type() reflect.Type {
	return reflect.TypeOf(map[string]string{})
}
func (MapStringString) NumArgs() (int, int)    { return 1, 1 }
func (MapStringString) CanBeValue(string) bool { return true }
func (MapStringString) Set(source reflect.Value, args []string) error {
	idx := strings.Index(args[0], "=")
	if idx < 0 {
		return fmt.Errorf("invalid config: %v", args[0])
	}
	elem := source
	if source.Kind() == reflect.Ptr {
		elem = source.Elem()
	}
	if elem.IsNil() {
		elem.Set(reflect.ValueOf(map[string]string{}))
	}
	key := args[0][:idx]
	value := args[0][idx+1:]
	elem.SetMapIndex(reflect.ValueOf(key), reflect.ValueOf(value))
	return nil
}
func (MapStringString) ArgName() string {
	return "key=value"
}
