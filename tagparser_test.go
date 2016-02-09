package flagly

import "testing"

func TestTagParser(t *testing.T) {
	s := `bson:"dfdf df df " json:"hello dfdf" xml:dfdf sd`
	st := StructTag(s)
	if st.Get("json") != "hello dfdf" {
		t.Fatal()
	}
	if st.Get("bson") != "dfdf df df " {
		t.Fatal()
	}
	if !st.Has("xml") {
		t.Fatal()
	}
	if st.Get("xml") != "dfdf" {
		t.Fatal()
	}

	st = StructTag(`quote optional default:"" required`)
	if !st.Has("optional") {
		t.Fatal()
	}
	if st.GetPtr("default") == nil {
		t.Fatal()
	}
	if st.GetPtr("kjkj") != nil {
		t.Fatal()
	}
	if st.GetName() != "quote" {
		t.Fatal()
	}
	if !st.Has("required") {
		t.Fatal()
	}

	st = StructTag(`"quote"`)
	if st.GetName() != "quote" {
		t.Fatal()
	}
}
