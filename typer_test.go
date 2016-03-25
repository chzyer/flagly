package flagly

import (
	"reflect"
	"testing"
)

func TestTyper(t *testing.T) {
	var mapStr map[string]string
	mapStrType := reflect.TypeOf(mapStr)
	typer, err := GetTyper(mapStrType)
	if err != nil {
		t.Fatal(err)
	}
	typer.Set(reflect.ValueOf(&mapStr), []string{"a=b"})
	if mapStr["a"] != "b" {
		t.Fatal("error")
	}
	typer.Set(reflect.ValueOf(&mapStr), []string{"b=c"})
	if mapStr["a"] != "b" || mapStr["b"] != "c" {
		t.Fatal("error")
	}
}
