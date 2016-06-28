package flagly

import (
	"fmt"
	"net"
	"reflect"
	"strconv"
	"strings"
	"time"
)

var (
	types    = map[string]Typer{}
	NilValue reflect.Value
)

func init() {
	RegisterAll(Bool{}, String{}, Duration{}, Int{}, Int64{}, IPNet{})
	Register(MapStringString{})
}

func getTypeName(t reflect.Type) string {
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	return t.String()
}

func RegisterAll(objs ...BaseTyperParser) {
	for _, obj := range objs {
		Register(ValueWrap{obj}, SliceWrap{obj})
	}
}

func Register(objs ...Typer) {
	for _, obj := range objs {
		t := obj.Type()
		if t.Kind() == reflect.Ptr {
			t = t.Elem()
		}
		types[getTypeName(t)] = obj
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
}
type BaseTypeArgNamer interface {
	ArgName() string
}
type BaseTypeCanBeValuer interface {
	CanBeValue(arg string) bool
}
type BaseTypeNumArgs interface {
	NumArgs() (int, int)
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

	for _, a := range args {
		arg, err := s.BaseTyperParser.ParseArgs([]string{a})
		if err != nil {
			return err
		}
		val := reflect.New(s.BaseTyperParser.Type())
		SetToSource(val, arg)
		source.Set(reflect.Append(source, val.Elem()))
	}
	return nil
}

func (s SliceWrap) CanBeValue(arg string) bool {
	if cbv, ok := s.BaseTyperParser.(BaseTypeCanBeValuer); ok {
		return cbv.CanBeValue(arg)
	}
	return true
}

func (s SliceWrap) NumArgs() (int, int) {
	if an, ok := s.BaseTyperParser.(BaseTypeNumArgs); ok {
		return an.NumArgs()
	}
	return 1, 1
}

type Typer interface {
	BaseTyper
	BaseTypeNumArgs
	BaseTypeCanBeValuer
	Set(source reflect.Value, args []string) error
}

type ValueWrap struct{ BaseTyperParser }

func SetToSource(source, val reflect.Value) {
	if source.Kind() == reflect.Ptr {
		if source.IsNil() {
			ptr := reflect.New(source.Type().Elem())
			source.Set(ptr)
		}
		if val.Kind() != reflect.Ptr {
			source = source.Elem()
		}
	}
	source.Set(val)
}

func (b ValueWrap) CanBeValue(arg string) bool {
	if cbv, ok := b.BaseTyperParser.(BaseTypeCanBeValuer); ok {
		return cbv.CanBeValue(arg)
	}
	return true
}

func (b ValueWrap) NumArgs() (int, int) {
	if an, ok := b.BaseTyperParser.(BaseTypeNumArgs); ok {
		return an.NumArgs()
	}
	return 1, 1
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

func (Int) Type() reflect.Type { return reflect.TypeOf(int(0)) }
func (Int) ArgName() string    { return "number" }
func (Int) ParseArgs(args []string) (reflect.Value, error) {
	val, err := strconv.Atoi(args[0])
	if err != nil {
		return NilValue, err
	}
	return reflect.ValueOf(val), nil
}

type Int64 struct{}

func (Int64) Type() reflect.Type { return reflect.TypeOf(int64(0)) }
func (Int64) ArgName() string    { return "number" }
func (Int64) ParseArgs(args []string) (reflect.Value, error) {
	val, err := strconv.ParseInt(args[0], 10, 64)
	if err != nil {
		return NilValue, err
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
		return NilValue, err
	}
	return reflect.ValueOf(a), nil
}

type IPNet struct{}

func (IPNet) Type() reflect.Type     { return reflect.TypeOf(&net.IPNet{}) }
func (IPNet) NumArgs() (int, int)    { return 1, 1 }
func (IPNet) CanBeValue(string) bool { return true }
func (IPNet) ParseArgs(args []string) (reflect.Value, error) {
	if idx := strings.Index(args[0], "/"); idx < 0 {
		args[0] += "/32"
	}

	_, ipnet, err := net.ParseCIDR(args[0])
	if err != nil {
		return NilValue, err
	}
	return reflect.ValueOf(ipnet), nil
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
