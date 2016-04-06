package flagly

import (
	"bytes"
	"strconv"
	"strings"
)

type StructTag string

func (st StructTag) Flagly() []string {
	sp := strings.Split(st.Get("flagly"), ",")
	return sp
}

func (st StructTag) FlaglyHas(name string) bool {
	for _, s := range st.Flagly() {
		if s == name {
			return true
		}
	}
	return false
}

func (st StructTag) GetPtr(name string) *string {
	if !st.Has(name) {
		return nil
	}
	ret := st.Get(name)
	return &ret
}

func (st StructTag) Get(name string) string {
	var idx int
	s := string(st)
	target := name + ":"
	for {
		idx = strings.Index(s, target)
		if idx >= 0 {
			if strings.Count(s, `"`)&1 == 0 {
				break
			} else {
				s = s[idx+1:]
			}
		} else {
			return ""
		}
	}

	// found
	quoted := false
	content := bytes.NewBuffer(nil)
	for i := idx + len(target); i < len(s); i++ {
		isQuoteChar := s[i] == '"'
		if !quoted && s[i] == ' ' {
			break
		}
		content.WriteByte(s[i])
		if isQuoteChar {
			if !quoted {
				quoted = true
			} else {
				quoted = false
				break
			}
		}
	}

	ret, err := strconv.Unquote(content.String())
	if err != nil {
		ret = content.String()
	}
	return ret
}

func (st StructTag) Has(name string) bool {
	s := string(st)
	var idx int
	for {
		idx = strings.Index(s, name)
		if idx < 0 {
			return false
		}
		if idx > 0 && s[idx-1] != ' ' {
			continue
		}
		if len(s) > idx+len(name) {
			ss := s[idx+len(name)]
			if ss == ' ' || ss == ':' {
				return true
			}
		} else {
			return true
		}
		s = s[idx+1:]
	}
}

func (st StructTag) GetName() string {
	if name := st.getName(); name != "" {
		return name
	}
	return st.Get("name")
}

func (st StructTag) getName() string {
	s := string(st)
	idx := strings.Index(s, " ")
	if idx >= 0 {
		s = s[:idx]
	}
	if strings.Contains(s, ":") {
		return ""
	}
	if len(s) >= 2 &&
		s[0] == '"' && s[len(s)-1] == '"' {
		tmpName, err := strconv.Unquote(s)
		if err == nil {
			s = tmpName
		}
	}
	s = strings.TrimSpace(s)
	return s
}
