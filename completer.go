package flagly

type Tree interface {
	GetName() string
	GetTreeChildren() []Tree
}

type HandlerCompleter struct {
	h Tree
}

func (hc *HandlerCompleter) DoSegment(seg [][]rune, n int) [][]rune {
	h := hc.h
main:
	for level := 0; level < len(seg)-1; {
		name := string(seg[level])
		children := h.GetTreeChildren()
		for _, child := range children {
			if child.GetName() == name {
				h = child
				level++
				continue main
			}
		}
		h = nil
		break
	}
	if h == nil {
		return nil
	}
	children := h.GetTreeChildren()
	ret := make([][]rune, len(children))
	for idx, child := range children {
		ret[idx] = []rune(child.GetName())
	}

	return ret
}

type stringTree struct {
	name string
}

func StringTree(name string) Tree {
	return stringTree{name}
}

func (n stringTree) GetName() string {
	return n.name
}

func (n stringTree) GetTreeChildren() []Tree {
	return nil
}
