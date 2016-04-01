package flagly

type HandlerCompleter struct {
	h *Handler
}

func (hc *HandlerCompleter) DoSegment(seg [][]rune, n int) [][]rune {
	h := hc.h
main:
	for level := 0; level < len(seg)-1; {
		name := string(seg[level])
		children := h.GetChildren()
		for _, child := range children {
			if child.Name == name {
				h = child
				level++
				continue main
			}
		}
	}
	children := h.GetChildren()
	ret := make([][]rune, len(children))
	for idx, child := range children {
		ret[idx] = []rune(child.Name)
	}

	return ret
}
