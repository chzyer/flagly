package flagly

import "reflect"

var (
	emptyFlaglyIniter FlaglyIniter
	FlaglyIniterName  = "FlaglyInit"
)

type FlaglyIniter interface {
	FlaglyInit()
}

func IsImplementIniter(t reflect.Type) bool {
	return IsImplemented(t, reflect.TypeOf(&emptyFlaglyIniter).Elem())
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
