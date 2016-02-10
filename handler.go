package flagly

import (
	"bytes"
	"fmt"
	"os"
	"reflect"
	"strings"
)

var (
	emptyError error
	emptyType  reflect.Type
	emptyValue reflect.Value
	IfaceError = reflect.TypeOf(&emptyError).Elem()
)

type Handler struct {
	Parent   *Handler
	Name     string
	Desc     string
	Children []*Handler

	Options       []*Option
	OptionType    reflect.Type
	handleFunc    reflect.Value
	onGetChildren func(*Handler) []*Handler
}

func NewHandler(name string) *Handler {
	h := &Handler{
		Name: name,
	}

	return h
}

func (h *Handler) ResetHandler() {
	h.Children = nil
}

func (h *Handler) AddHandler(child *Handler) {
	child.Parent = h
	h.Children = append(h.Children, child)
}

func (h *Handler) SetGetChildren(f func(*Handler) []*Handler) {
	h.onGetChildren = f
}

func (h *Handler) SetOptionType(option reflect.Type) error {
	op := option
	if op.Kind() == reflect.Ptr {
		op = op.Elem()
	}
	if op.Kind() == reflect.Struct {
		if op.String() != handlerPkgPath {
			h.OptionType = option
		}
	}
	if h.OptionType != nil {
		var err error
		h.Options, err = h.parseOption()
		if err != nil {
			return err
		}
	}
	return nil
}

// only func is accepted
// 1. func() error
// 2. func(*struct) error
// 3. func(*struct, *flagly.Handler) error
// 4. func(*flagly.Handler) error
func (h *Handler) SetHandleFunc(obj interface{}) error {
	return h.setHandleFunc(reflect.ValueOf(obj))
}

func (h *Handler) setHandleFunc(funcValue reflect.Value) error {
	if funcValue.Kind() != reflect.Func {
		return ErrMustAFunc
	}
	t := funcValue.Type()
	if t.NumOut() != 1 || !t.Out(0).Implements(IfaceError) {
		return ErrFuncOutMustAError
	}
	if t.NumIn() >= 1 {
		if err := h.SetOptionType(t.In(0)); err != nil {
			return err
		}
	}
	h.handleFunc = funcValue
	return nil
}

func (h *Handler) findHandleFunc(t reflect.Type) *reflect.Method {
	for i := 0; i < t.NumMethod(); i++ {
		method := t.Method(i)
		if method.Name == flaglyHandle {
			return &method
		}
	}
	return nil
}

func (h *Handler) Compile(t reflect.Type) error {
	method := h.findHandleFunc(t)
	if method != nil {
		if err := h.setHandleFunc(method.Func); err != nil {
			return err
		}
	} else {
		if err := h.SetOptionType(t); err != nil {
			return err
		}
	}

	if IsImplementDescer(h.OptionType) {
		op := h.OptionType
		if op.Kind() == reflect.Ptr {
			op = op.Elem()
		}
		value := reflect.New(op)
		h.Desc = value.Interface().(FlaglyDescer).FlaglyDesc()
	}

	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if StructTag(field.Tag).GetName() == flaglyHandler {
			subh := NewHandler(strings.ToLower(field.Name))
			if err := subh.Compile(field.Type); err != nil {
				return err
			}
			h.AddHandler(subh)
		}
	}

	return nil
}

func (h *Handler) HasDesc() bool {
	return h.Desc != ""
}

func (h *Handler) writeDesc(buf *bytes.Buffer) {
	prefix := "    "
	space := 20
	if len(h.Name) > space {
		buf.WriteString(prefix + h.Name + "\n")
		if h.HasDesc() {
			buf.WriteString(strings.Repeat(" ", space) + h.Desc + "\n")
		}
	} else {
		indent := strings.Repeat(" ", space-len(h.Name))
		buf.WriteString(prefix + h.Name + indent + h.Desc + "\n")
	}
}

func (h *Handler) GetChildren() []*Handler {
	if h.onGetChildren != nil {
		ch := h.onGetChildren(h)
		for _, c := range ch {
			h.AddHandler(c)
		}
		return ch
	}
	return h.Children
}

func (h *Handler) GetHandler(name string) *Handler {
	for _, ch := range h.GetChildren() {
		if ch.Name == name {
			return ch
		}
	}
	return nil
}

func (h *Handler) Usage(prefix string) string {
	buf := bytes.NewBuffer(nil)
	err := h.usage(buf, prefix)
	if err != nil {
		println(err.Error())
		os.Exit(2)
	}
	return buf.String()
}

func (h *Handler) parseOption() ([]*Option, error) {
	if h.OptionType == emptyType {
		return nil, nil
	}
	ops, err := ParseStructToOptions(h.OptionType)
	hasHelp := false
	for _, op := range ops {
		if op.Name == "-h" {
			hasHelp = true
			break
		}
	}
	if !hasHelp {
		ops = append(ops, NewHelpFlag())
	}
	return ops, err
}

func (h *Handler) findOption(name string) int {
	for idx, op := range h.Options {
		if op.Name == name {
			return idx
		}
	}
	return -1
}

func (h *Handler) parseToStruct(v reflect.Value, args []string) ([]string, error) {
	tokens := make([][]string, len(h.Options))
	idx := 0
	for ; idx < len(args); idx++ {
		arg := args[idx]
		if strings.HasPrefix(arg, "-") {
			// TODO: flag name can't be -
			if arg == "--" {
				args = args[idx+1:]
				break
			}
			opIdx := h.findOption(arg[1:])
			if opIdx < 0 {
				continue
			}
			op := h.Options[opIdx]
			if op.ShowUsage {
				return args, ErrShowUsage
			}
			min, max := op.Typer.NumArgs()
			subArgs := make([]string, 0, max)
			for i := idx + 1; i <= idx+max && i < len(args); i++ {
				if !op.Typer.CanBeValue(args[i]) {
					break
				}
				subArgs = append(subArgs, args[i])
			}
			if len(subArgs) < min {
				return args, fmt.Errorf("args missing")
			}
			idx += len(subArgs)
			tokens[opIdx] = subArgs
			continue
		}
		break
	}

	args = args[idx:]

	for idx, op := range h.Options {
		if op.IsArg() {
			if op.ArgIdx == -1 {
				op.BindTo(v, args)
			} else if op.ArgIdx >= len(args) {
				if op.HasDefault() {
					op.BindTo(v, nil)
				} else {
					// do nothing
				}
			} else {
				op.BindTo(v, args[op.ArgIdx:op.ArgIdx+1])
			}
		} else if op.IsFlag() {
			op.BindTo(v, tokens[idx])
		} else {
			return args, fmt.Errorf("invalid option type: %v", op.Type)
		}
	}
	return args, nil
}

func (h *Handler) bindStackToStruct(stack []reflect.Value, value reflect.Value) {
	if value.Kind() == reflect.Ptr {
		value = value.Elem()
	}
	t := value.Type()
	for i := 0; i < value.NumField(); i++ {
		field := t.Field(i)
		if StructTag(field.Tag).GetName() == flaglyParentName {
			for _, s := range stack {
				if s.Type().String() == field.Type.String() {
					value.Field(i).Set(s)
				}
			}
		}
	}
}

func (h *Handler) Call(stack []reflect.Value, args []string) error {
	if h.handleFunc.IsValid() {
		t := h.handleFunc.Type()
		numIn := t.NumIn()
		ins := make([]reflect.Value, numIn)
		for i := 0; i < numIn; i++ {
			tIn := t.In(i)
			switch tIn.String() {
			case "*" + handlerPkgPath:
				// TODO: must be a pointer
				ins[i] = reflect.ValueOf(h)
			default:
				ins[i] = reflect.Zero(t.In(i))
			}
		}
		if len(ins) > 0 {
			opType := t.In(0)
			if opType.String() != "*"+handlerPkgPath {
				if opType.Kind() == reflect.Ptr {
					ins[0] = reflect.New(opType.Elem())
				}
				if _, err := h.parseToStruct(ins[0], args); err != nil {
					return err
				}
				h.bindStackToStruct(stack, ins[0])
			}
		}
		// first argument is a struct
		out := h.handleFunc.Call(ins)
		if err, ok := out[0].Interface().(error); ok {
			return err
		}
		return nil
	} else {
		// show usage
		return ErrShowUsage
	}
}

func (h *Handler) getChildNames() (names []string) {
	for _, n := range h.GetChildren() {
		names = append(names, n.Name)
	}
	return names
}

func (h *Handler) Run(stack *[]reflect.Value, args []string) (err error) {
	defer func() {
		if e := IsShowUsage(err); e != nil {
			err = e.Trace(h)
		}
	}()
	runed := false
	var value reflect.Value
	if h.OptionType != nil {
		t := h.OptionType
		if t.Kind() == reflect.Ptr {
			t = t.Elem()
		}
		value = reflect.New(t)
		args, err = h.parseToStruct(value, args)
		if err != nil {
			return err
		}
	}
	*stack = append(*stack, value)

	if len(args) > 0 {
		for _, ch := range h.GetChildren() {
			if args[0] == ch.Name {
				err = ch.Run(stack, args[1:])
				runed = true
				break
			}
		}
	}
	if !runed {
		err = h.Call(*stack, args)
	}
	return err
}

func (h *Handler) String() string {
	return fmt.Sprintf("%+v", *h)
}

func (h *Handler) HasFlagOptions() bool {
	if len(h.Options) > 0 {
		for _, op := range h.Options {
			if op.IsFlag() {
				return true
			}
		}
	}
	return false
}

func (h *Handler) UsagePrefix() string {
	buf := bytes.NewBuffer(nil)
	buf.WriteString(h.Name)
	if h.HasFlagOptions() {
		buf.WriteString(" [" + h.Name + " option]")
	}
	return buf.String()
}

func (h *Handler) HasArgOptions() bool {
	if len(h.Options) > 0 {
		for _, op := range h.Options {
			if op.IsArg() {
				return true
			}
		}
	}
	return false
}

func (h *Handler) usage(buf *bytes.Buffer, prefix string) error {
	if prefix != "" {
		prefix += " "
	}

	hasFlags := h.HasFlagOptions()
	hasArgs := h.HasArgOptions()
	children := h.GetChildren()
	hasCommands := len(children) > 0

	buf.WriteString("usage: " + prefix + h.Name)
	if hasFlags {
		buf.WriteString(" [option]")
	}
	if hasCommands {
		buf.WriteString(" <command>")
	}
	if hasArgs {
		if hasFlags {
			buf.WriteString(" [--]")
		}
		for _, op := range h.Options {
			if op.IsArg() {
				buf.WriteString(" ")
				if op.HasDefault() {
					buf.WriteString("[")
				}
				buf.WriteString("<" + op.Name + ">")
				if op.HasDefault() {
					buf.WriteString("]")
				}
			}
		}
	}
	buf.WriteString("\n")

	if hasFlags {
		buf.WriteString(h.usageOptions("options"))
	}

	if hasCommands {
		buf.WriteString("\ncommands:\n")
		for _, ch := range children {
			ch.writeDesc(buf)
		}
	}

	return nil

}

func (h *Handler) usageOptions(name string) string {
	buf := bytes.NewBuffer(nil)
	buf.WriteString("\n" + name + ":\n")
	for _, op := range h.Options {
		if op.IsFlag() {
			op.usage(buf)
			buf.WriteString("\n")
		}
	}
	return buf.String()
}

func (h *Handler) Bind(ptr reflect.Value, args []string) (err error) {
	if ptr.Kind() != reflect.Ptr {
		return ErrMustAPtrToStruct
	}
	defer func() {
		if e := IsShowUsage(err); e != nil {
			err = e.Trace(h)
		}
	}()
	if h.OptionType != nil {
		t := h.OptionType
		if t.Kind() == reflect.Ptr {
			t = t.Elem()
		}
		value := reflect.New(t)
		_, err = h.parseToStruct(value, args)
		if err != nil {
			return err
		}

		ptr.Elem().Set(value.Elem())
	}
	return nil
}
