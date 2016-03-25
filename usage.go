package flagly

import (
	"fmt"
	"strings"
)

var (
	ErrShowUsage error = showUsageError{}
)

type showUsageError struct {
	info     string
	handlers []*Handler
}

func (s showUsageError) Error() string {
	if s.info != "" {
		usage := s.Usage()
		if usage != "" {
			return s.info + "\n\n" + usage
		}
		return s.info

	}
	return s.Usage()
}

func (s *showUsageError) Trace(h *Handler) *showUsageError {
	s.handlers = append(s.handlers, h)
	return s
}

func (s *showUsageError) Usage() string {
	return ShowUsage(s.handlers)
}

func Errorf(format string, obj ...interface{}) error {
	return Error(fmt.Sprintf(format, obj...))
}

func Error(info string) error {
	return &showUsageError{
		info: info,
	}
}

func ShowUsage(hs []*Handler) string {
	prefix := ""
	for i := len(hs) - 1; i > 0; i-- {
		prefix += hs[i].UsagePrefix() + " "
	}
	prefix = strings.TrimSpace(prefix)
	if len(hs) > 0 {
		h := hs[0]
		usage := h.Usage(prefix)
		for i := 1; i < len(hs); i++ {
			if hs[i].HasFlagOptions() {
				usage += hs[i].usageOptions(hs[i].Name + " options")
			}
		}
		return usage
	} else {
		return ""
	}
}
