package hidden_fuzzer

type Request struct {
	Method  string
	Host    string
	URL     string
	Headers map[string]string
}
