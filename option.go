package flagly

import (
	"bytes"
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

type OptionType int

func (t OptionType) String() string {
	switch t {
	case FlagOption:
		return "flag"
	case ArgOption:
		return "arg"
	default:
		return "<unknown>"
	}
}

const (
	FlagOption OptionType = iota
	ArgOption
)

type Option struct {
	Index    int
	Name     string
	LongName string
	Type     OptionType
	BindType reflect.Type
	Typer    Typer
	Desc     string
	Default  *string
	ArgName  *string
	ArgIdx   int
}

func NewFlag(name string, bind reflect.Type) *Option {
	return &Option{
		Name:     name,
		BindType: bind,
		Type:     FlagOption,
	}
}

func NewArg(name string, idx int, bind reflect.Type) *Option {
	return &Option{
		Name:     name,
		Type:     ArgOption,
		BindType: bind,
		ArgIdx:   idx,
	}
}

func (o *Option) BindTo(value reflect.Value, args []string) {
	f := value.Elem().Field(o.Index)
	if args == nil {
		if o.HasDefault() {
			o.Typer.Set(f, []string{*o.Default})
		}
	} else {
		o.Typer.Set(f, args)
	}
}

func (o *Option) HasArgName() bool {
	return o.ArgName != nil
}

func (o *Option) HasDefault() bool {
	return o.Default != nil
}

func (o *Option) IsFlag() bool {
	return o.Type == FlagOption
}

func (o *Option) IsArg() bool {
	return o.Type == ArgOption
}

func (o *Option) usage(buf *bytes.Buffer) {
	b := bytes.NewBuffer(nil)
	b.WriteString("    ")
	length := 4 + 20
	if o.IsFlag() {
		b.WriteString("-" + o.Name)
		min, _ := o.Typer.NumArgs()
		if min > 0 && o.HasArgName() {
			if o.Default != nil {
				b.WriteString("[=")
			} else {
				b.WriteString(" ")
			}
			b.WriteString("<" + *o.ArgName)
			if o.HasDefault() {
				b.WriteString("=" + *o.Default)
			}
			b.WriteString(">")
			if o.Default != nil {
				b.WriteString("]")
			}
		}
		if b.Len() > length {
			b.WriteString("\n" + strings.Repeat(" ", length))
		} else {
			b.WriteString(strings.Repeat(" ", length-b.Len()))
		}
		b.WriteString(o.Desc)
	}
	b.WriteTo(buf)
}

func GetNameFromTag(tag StructTag) (name, longName string) {
	name = tag.GetName()
	if IsWrapBy(name, "[]") {
		longName = tag.Get("name")
	} else {
		longName = tag.Get("long")
	}
	return
}

func IsWrapBy(s, ch2 string) bool {
	return len(s) >= 2 && s[0] == ch2[0] && s[len(s)-1] == ch2[1]
}

func GetIdxInArray(s string) int {
	s = s[1 : len(s)-1]
	idx, err := strconv.Atoi(s)
	if err != nil {
		return -1
	}
	return idx
}

func GetMethod(s reflect.Value, name string) reflect.Value {
	method := s.MethodByName(name)
	if !method.IsValid() {
		method = s.Elem().MethodByName(name)
	}
	return method
}

func ParseStructToOptions(t reflect.Type) (ret []*Option, err error) {
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	descIdx := make(map[int]string)
	if IsImplementIniter(t) {
		value := reflect.New(t)
		GetMethod(value, FlaglyIniterName).Call(nil)
		descMap := getDescMap()
		elem := value.Elem()
		for i := 0; i < elem.NumField(); i++ {
			desc, ok := descMap[elem.Field(i).UnsafeAddr()]
			if !ok {
				continue
			}
			descIdx[i] = desc
		}
	}

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		tag := StructTag(field.Tag)
		short, long := GetNameFromTag(tag)
		if short == "" {
			short = strings.ToLower(field.Name)
		}
		var op *Option

		if IsWrapBy(short, "[]") {
			if long == "" {
				long = strings.ToLower(field.Name)
			}
			op = NewArg(long, GetIdxInArray(short), field.Type)
		} else {
			op = NewFlag(short, field.Type)
			op.LongName = long
		}
		if op.Name == "-" {
			return nil, fmt.Errorf(`name "-" is not allowed`)
		}
		op.Index = i

		op.Default = tag.GetPtr("default")
		typer, err := GetTyper(op.BindType)
		if err != nil {
			return nil, err
		}
		op.Typer = typer
		if namer, ok := typer.(BaseTypeArgNamer); ok {
			argName := namer.ArgName()
			if argName != "" {
				op.ArgName = &argName
			}
		}

		if argName := tag.GetPtr("arg"); argName != nil {
			op.ArgName = argName
		}
		op.Desc = tag.Get("desc")
		if desc, ok := descIdx[i]; ok {
			op.Desc = desc
		}

		ret = append(ret, op)
	}
	return
}

func (o *Option) String() string {
	return fmt.Sprintf("%v", *o)
}
