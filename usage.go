package flagly

import "strings"

var (
	ErrShowUsage error = showUsageError{}
)

type showUsageError struct {
	handlers []*Handler
}

func (s showUsageError) Error() string {
	return s.Usage()
}

func (s *showUsageError) Trace(h *Handler) *showUsageError {
	s.handlers = append(s.handlers, h)
	return s
}

func (s *showUsageError) Usage() string {
	return ShowUsage(s.handlers)
}

func ShowUsage(hs []*Handler) string {
	prefix := ""
	for i := len(hs) - 1; i > 0; i-- {
		prefix += hs[i].UsagePrefix() + " "
	}
	prefix = strings.TrimSpace(prefix)
	h := hs[0]
	usage := h.Usage(prefix)
	for i := 1; i < len(hs); i++ {
		if hs[i].HasFlagOptions() {
			usage += hs[i].usageOptions(hs[i].Name + " options")
		}
	}
	return usage
}
