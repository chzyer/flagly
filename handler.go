package flagly

import (
	"bytes"
	"fmt"
	"os"
	"reflect"
	"strings"
)

var (
	EmptyError  error
	EmptyType   reflect.Type
	EmptyValue  reflect.Value
	IfaceError  = reflect.TypeOf(&EmptyError).Elem()
	HandlerType = reflect.TypeOf(new(Handler))
)

type Handler struct {
	Parent   *Handler
	Name     string
	Desc     string
	Children []*Handler

	context       map[string]reflect.Value
	lambdaMap     map[string]func() []string
	Options       []*Option
	OptionType    reflect.Type
	handleFunc    reflect.Value
	onGetChildren func(*Handler) []*Handler
	onExit        func()
}

func NewHandler(name string) *Handler {
	h := &Handler{
		Name:      name,
		context:   make(map[string]reflect.Value),
		lambdaMap: make(map[string]func() []string),
	}
	return h
}

func (h *Handler) GetRoot() *Handler {
	if h.Parent != nil {
		return h.Parent.GetRoot()
	}
	return h
}

func (h *Handler) GetName() string {
	return h.Name
}

func (h *Handler) findArgOption() *Option {
	for _, op := range h.Options {
		if op.IsArg() {
			return op
		}
	}
	return nil
}

func (h *Handler) GetTreeChildren() []Tree {
	children := h.GetChildren()
	if len(children) == 0 {
		argOp := h.findArgOption()
		if argOp != nil {
			return argOp.GetTree(h.lambdaMap)
		}
	}
	ret := make([]Tree, len(children))
	for idx, ch := range children {
		ret[idx] = ch
	}
	return ret
}

func (h *Handler) SetOnExit(f func()) {
	h.onExit = f
}

func (h *Handler) ResetHandler() {
	h.Children = nil
}

// combine NewHander(name).SetHanderFunc()/AddHandler(subHandler)
func (h *Handler) AddSubHandler(name string, function interface{}) *Handler {
	subHandler := NewHandler(name)
	subHandler.SetHandleFunc(function)
	h.AddHandler(subHandler)
	return subHandler
}

func (h *Handler) copyContext() {
	for _, ch := range h.Children {
		ch.context = h.context
		ch.lambdaMap = h.lambdaMap
		ch.copyContext()
	}
}

func (h *Handler) Lambda(name string, fn func() []string) {
	h.lambdaMap[name] = fn
}

func (h *Handler) AddHandler(child *Handler) {
	child.Parent = h
	h.Children = append(h.Children, child)
	h.copyContext()
	child.EnsureHelpOption()
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

		if IsImplementDescer(h.OptionType) {
			op := h.OptionType
			if op.Kind() == reflect.Ptr {
				op = op.Elem()
			}
			value := reflect.New(op)
			h.Desc = value.Interface().(FlaglyDescer).FlaglyDesc()
		}
	}
	return nil
}

// only func is accepted
// 1. func() error
// 2. func(*struct) error
// 3. func(*struct, *flagly.Handler) error
// 4. func(*flagly.Handler) error
// 5. func(Handler, Context) error
func (h *Handler) SetHandleFunc(obj interface{}) {
	err := h.setHandleFunc(reflect.ValueOf(obj))
	if err != nil {
		panic(err)
	}
}

func (h *Handler) setHandleFunc(funcValue reflect.Value) error {
	if funcValue.Kind() != reflect.Func {
		return ErrMustAFunc
	}
	t := funcValue.Type()
	if t.NumOut() != 1 || !t.Out(0).Implements(IfaceError) {
		if t.NumOut() != 0 {
			return fmt.Errorf(ErrFuncOutMustAError.Error(), h.Name)
		}
	}
	if t.NumIn() >= 1 {
		if err := h.SetOptionType(t.In(0)); err != nil {
			return err
		}
	}
	h.handleFunc = funcValue
	return nil
}

func (h *Handler) findEnterFunc(t reflect.Type) *reflect.Method {
	method, ok := t.MethodByName(flaglyEnter)
	if ok {
		return &method
	}
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

func (h *Handler) EnsureHelpOption() {
	h.Options = h.tryToAddHelpOption(h.Options)
}

func (h *Handler) tryToAddHelpOption(ops []*Option) []*Option {
	if len(h.GetChildren()) == 0 {
		hasHelp := false
		for _, op := range ops {
			if op.Name == "h" {
				hasHelp = true
				break
			}
		}
		if !hasHelp {
			ops = append(ops, NewHelpFlag())
		}
	}
	return ops
}

func (h *Handler) CompileIface(obj interface{}) error {
	return h.Compile(reflect.TypeOf(obj))
}

func (h *Handler) log(obj interface{}) {
	println(fmt.Sprintf("[handler:%v] %v", h.Name, obj))
}

func (h *Handler) Context(obj interface{}) {
	value := reflect.ValueOf(obj)
	typ := value.Type()
	if typ.Kind() == reflect.Ptr {
		h.context[typ.String()] = value
		h.context[typ.Elem().String()] = value
	} else {
		h.context[typ.String()] = value
		h.context[reflect.PtrTo(typ).String()] = value
	}
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

	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		tag := StructTag(field.Tag)
		if tag.FlaglyHas("handler") {
			name := tag.GetName()
			if name == "" {
				name = strings.ToLower(field.Name)
			}
			subh := NewHandler(name)
			if err := subh.Compile(field.Type); err != nil {
				return err
			}
			h.AddHandler(subh)
		}
	}

	h.Options = h.tryToAddHelpOption(h.Options)

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
	if h.OptionType == EmptyType {
		return nil, nil
	}
	ops, err := ParseStructToOptions(h, h.OptionType)
	return ops, err
}

func (h *Handler) GetOptionNames() []string {
	ret := make([]string, 0, len(h.Options))
	for _, op := range h.Options {
		ret = append(ret, op.Name)
	}
	return ret
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
		var err error
		if op.IsArg() {
			if op.ArgIdx == -1 {
				err = op.BindTo(v, args)
			} else if op.ArgIdx >= len(args) {
				if op.HasDefault() {
					err = op.BindTo(v, nil)
				} else {
					// do nothing
				}
			} else {
				err = op.BindTo(v, args[op.ArgIdx:op.ArgIdx+1])
			}
		} else if op.IsFlag() {
			err = op.BindTo(v, tokens[idx])
		} else {
			return args, fmt.Errorf("invalid option type: %v", op.Type)
		}
		if err != nil {
			return nil, err
		}
	}
	if IsImplementVerifier(v.Type()) {
		if err := v.Interface().(FlaglyVerifier).FlaglyVerify(); err != nil {
			return args, Error(err.Error())
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
		tag := StructTag(field.Tag)
		if tag.GetName() == flaglyParentName ||
			tag.FlaglyHas("parent") {
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
			if i == 0 {
				ins[0] = stack[len(stack)-1]
				if tIn.Kind() != reflect.Ptr {
					ins[0] = ins[0].Elem()
				}
				continue
			}
			switch tIn.String() {
			case "*" + handlerPkgPath:
				// TODO: must be a pointer
				ins[i] = reflect.ValueOf(h)
			default:
				if val, ok := h.context[tIn.String()]; ok {
					if val.Type().String() != tIn.String() {
						if strings.HasPrefix(tIn.String(), "*") {
							// want a pointer
							// todo, make a pointer
							val = reflect.New(tIn)
						} else {
							val = val.Elem()
						}
					}
					ins[i] = val
				} else {
					ins[i] = reflect.Zero(t.In(i))
				}
			}
		}
		// first argument is a struct
		out := h.handleFunc.Call(ins)
		if len(out) != 0 {
			if err, ok := out[0].Interface().(error); ok {
				return err
			}
		}
		return nil
	} else {
		// show usage
		return ErrShowUsage
	}
}

func (h *Handler) GetCommands() []string {
	return h.getChildNames()
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
		enter := h.findEnterFunc(h.OptionType)
		if enter != nil {
			args := make([]reflect.Value, 1)
			args[0] = value
			enter.Func.Call(args)
		}
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

	if h.Name != "" {
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
	}

	if h.Desc != "" {
		buf.WriteString(h.Desc)
		buf.WriteString("\n")
	}

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
		op.usage(buf)
		buf.WriteString("\n")
	}
	return buf.String()
}

func (h *Handler) Close() {
	if h.onExit != nil {
		h.onExit()
	}
	for _, ch := range h.GetChildren() {
		ch.Close()
	}
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
