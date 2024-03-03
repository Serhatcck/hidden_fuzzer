package hidden_fuzzer

import "time"

type Response struct {
	URL            string
	StatusCode     int
	Headers        map[string][]string
	DataForSimilar string
	Body           string
	ContentLength  int64
	ContentType    string
	Time           time.Duration
	Request        Request
	Error          error
}
