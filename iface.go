package flagly

import "reflect"

var (
	emptyFlaglyIniter FlaglyIniter
	emptyFlaglyDescer FlaglyDescer
	FlaglyIniterName  = "FlaglyInit"
	flaglyHandle      = "FlaglyHandle"
)

type FlaglyIniter interface {
	FlaglyInit()
}

type FlaglyDescer interface {
	FlaglyDesc() string
}

func IsImplementIniter(t reflect.Type) bool {
	return IsImplemented(t, reflect.TypeOf(&emptyFlaglyIniter).Elem())
}

func IsImplementDescer(t reflect.Type) bool {
	return IsImplemented(t, reflect.TypeOf(&emptyFlaglyDescer).Elem())
}

func IsImplemented(t, target reflect.Type) bool {
	if t.Implements(target) {
		return true
	}
	if t.Kind() == reflect.Struct {
		return reflect.PtrTo(t).Implements(target)
	} else if t.Kind() == reflect.Ptr {
		return t.Elem().Implements(target)
	}
	return false
}
