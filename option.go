package flagly

import (
	"bytes"
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

var (
	flaglyPrefix     = "flagly"
	flaglyParentName = "flaglyParent"
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
	Index     int
	Name      string
	LongName  string
	Type      OptionType
	BindType  reflect.Type
	Typer     Typer
	Desc      string
	Default   *string
	ArgName   *string
	ArgIdx    int
	ShowUsage bool
	Tag       StructTag
}

func NewHelpFlag() *Option {
	op, err := NewFlag("h", reflect.TypeOf(true))
	if err != nil {
		panic(err)
	}
	op.ShowUsage = true
	op.Desc = "show help"
	return op
}

func NewFlag(name string, bind reflect.Type) (*Option, error) {
	op := &Option{
		Index:    -1,
		Name:     name,
		BindType: bind,
		Type:     FlagOption,
	}
	if err := op.init(); err != nil {
		return nil, err
	}
	return op, nil
}

func NewArg(name string, idx int, bind reflect.Type) (*Option, error) {
	op := &Option{
		Index:    -1,
		Name:     name,
		Type:     ArgOption,
		BindType: bind,
		ArgIdx:   idx,
	}
	if err := op.init(); err != nil {
		return nil, err
	}
	return op, nil
}

func (o *Option) init() error {
	typer, err := GetTyper(o.BindType)
	if err != nil {
		return err
	}
	o.Typer = typer
	return nil
}

func (o *Option) GetTree(lambdaMap map[string]func() []string) []Tree {
	var candidates []string
	if selectCall := o.Tag.Get("selectCall"); selectCall != "" {
		if fn := lambdaMap[selectCall]; fn != nil {
			candidates = fn()
		}
	}
	if tags := o.Tag.Get("select"); tags != "" {
		candidates = strings.Split(tags, ",")
	}
	trees := make([]Tree, len(candidates))
	for idx, tag := range candidates {
		trees[idx] = StringTree(tag)
	}
	return trees
}

func (o *Option) BindTo(value reflect.Value, args []string) error {
	if o.Index < 0 {
		return nil
	}
	f := value.Elem().Field(o.Index)
	if args == nil {
		if o.HasDefault() {
			return o.Typer.Set(f, []string{*o.Default})
		}
	} else {
		return o.Typer.Set(f, args)
	}
	return nil
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

func (o *Option) HasDesc() bool {
	return o.Desc != ""
}

func (o *Option) usage(buf *bytes.Buffer) {
	b := bytes.NewBuffer(nil)
	b.WriteString("    ")
	length := 4 + 20
	if o.IsFlag() {
		b.WriteString("-" + o.Name)
		min, _ := o.Typer.NumArgs()

		if min > 0 {
			if o.HasArgName() {
				if o.HasDefault() {
					b.WriteString(fmt.Sprintf(" <%v=%v>", *o.ArgName, *o.Default))
				} else {
					b.WriteString(fmt.Sprintf(" <%v>", *o.ArgName))
				}
			} else if o.HasDefault() {
				b.WriteString(fmt.Sprintf(` "%v"`, *o.Default))
			} else {

			}
		}
	} else if o.IsArg() {
		b.WriteString(o.Name)
		if o.Tag.Get("select") != "" {
			o.Desc = o.Tag.Get("select")
		}
	}

	if o.HasDesc() {
		if b.Len() > length {
			b.WriteString("\n" + strings.Repeat(" ", length))
		} else {
			b.WriteString(strings.Repeat(" ", length-b.Len()))
		}
		b.WriteString(o.Desc)
	}

	b.WriteTo(buf)
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

func ParseStructToOptions(h *Handler, t reflect.Type) (ret []*Option, err error) {
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	descIdx := make(map[int]string)

	value := reflect.New(t)
	method := GetMethod(value, FlaglyIniterName)
	if method.IsValid() {
		methodType := method.Type()
		args := make([]reflect.Value, methodType.NumIn())
		for i := 0; i < methodType.NumIn(); i++ {
			typ := methodType.In(i)
			if typ.String() == HandlerType.String() {
				args[i] = reflect.ValueOf(h)
			} else {
				args[i] = reflect.Zero(typ)
			}
		}
		method.Call(args)
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
		if strings.HasPrefix(tag.GetName(), flaglyPrefix) || tag.Has("flagly") {
			continue
		}

		name := tag.GetName()
		if name == "" {
			name = strings.ToLower(field.Name)
		} else if name == "-" {
			continue
		}
		var op *Option

		if IsWrapBy(tag.Get("type"), "[]") {
			op, err = NewArg(name, GetIdxInArray(tag.Get("type")), field.Type)
		} else {
			op, err = NewFlag(name, field.Type)
		}
		if err != nil {
			return nil, err
		}
		op.Tag = tag

		if op.Name == "-" {
			return nil, fmt.Errorf(`name "-" is not allowed`)
		}
		op.Index = i

		op.Default = tag.GetPtr("default")
		if namer, ok := op.Typer.(BaseTypeArgNamer); ok {
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
