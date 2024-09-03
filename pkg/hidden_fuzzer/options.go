package hidden_fuzzer

import (
	"fmt"
	"strings"
)

type headerFlags map[string]string

func (h *headerFlags) String() string {
	return fmt.Sprint(*h)
}

func (h *headerFlags) Set(value string) error {
	if *h == nil {
		*h = make(map[string]string)
	}
	parts := strings.SplitN(value, ":", 2)
	if len(parts) != 2 {
		return fmt.Errorf("invalid header format, expecting key:value, got %s", value)
	}
	key := strings.TrimSpace(parts[0])
	val := strings.TrimSpace(parts[1])
	(*h)[key] = val
	return nil
}

type Options struct {
	Url                 string
	Wordlist            string
	Extensions          string
	Headers             headerFlags
	Method              string
	Threads             int
	FailureConter       int
	DuplicateCounter    int
	RedirectConter      int
	Silent              bool
	FailureCheckTimeout int
	TimeOut             int
	Depth               int
	RateLimit           int
}
