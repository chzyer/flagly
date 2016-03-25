package flagly

import "testing"

func TestTagParser(t *testing.T) {
	s := `bson:"dfdf df df " json:"hello dfdf" xml:dfdf sd`
	st := StructTag(s)
	if st.Get("json") != "hello dfdf" {
		t.Fatal("error")
	}
	if st.Get("bson") != "dfdf df df " {
		t.Fatal("error")
	}
	if !st.Has("xml") {
		t.Fatal("error")
	}
	if st.Get("xml") != "dfdf" {
		t.Fatal("error")
	}

	st = StructTag(`quote optional default:"" required`)
	if !st.Has("optional") {
		t.Fatal("error")
	}
	if st.GetPtr("default") == nil {
		t.Fatal("error")
	}
	if st.GetPtr("kjkj") != nil {
		t.Fatal("error")
	}
	if st.GetName() != "quote" {
		t.Fatal("error")
	}
	if !st.Has("required") {
		t.Fatal("error")
	}

	st = StructTag(`"quote"`)
	if st.GetName() != "quote" {
		t.Fatal("error")
	}
}
