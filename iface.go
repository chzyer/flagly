package flagly

import "reflect"

var (
	emptyFlaglyIniter   = reflect.TypeOf(new(FlaglyIniter)).Elem()
	emptyFlaglyDescer   = reflect.TypeOf(new(FlaglyDescer)).Elem()
	emptyFlaglyVerifier = reflect.TypeOf(new(FlaglyVerifier)).Elem()
	FlaglyIniterName    = "FlaglyInit"
	flaglyHandle        = "FlaglyHandle"
	flaglyEnter         = "FlaglyEnter"
)

type FlaglyIniter interface {
	FlaglyInit()
}

type FlaglyDescer interface {
	FlaglyDesc() string
}

type FlaglyVerifier interface {
	FlaglyVerify() error
}

func IsImplementIniter(t reflect.Type) bool {
	return IsImplemented(t, emptyFlaglyIniter)
}

func IsImplementDescer(t reflect.Type) bool {
	return IsImplemented(t, emptyFlaglyDescer)
}

func IsImplementVerifier(t reflect.Type) bool {
	return IsImplemented(t, emptyFlaglyVerifier)
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
