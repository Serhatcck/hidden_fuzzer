package hidden_fuzzer

import (
	"context"
	"errors"
	"net/url"
	"strings"
)

type Config struct {
	Context             context.Context
	Target              string
	Url                 *url.URL
	Silent              bool
	Wordlist            []string
	Extensions          []string
	Threads             int
	Headers             map[string]string
	FailureCheckTimeout int
	TimeOut             int
	Method              string
	FailureCounter      int
	DuplicateCounter    int
	RedirectCounter     int
	Depth               int
	RateLimit           int
	UseRateLimit        bool
}

func (c *Config) Build(options Options) error {
	c.Headers = make(map[string]string)

	//headers
	for key, value := range options.Headers {
		c.Headers[key] = value
	}
	if c.Headers["User-Agent"] == "" {
		c.Headers["User-Agent"] = "Chrome"
	}

	//url
	url, err := getURL(options.Url)
	if err {
		return errors.New("given url isn't url")
	}
	c.Url = url

	//wordlist
	wordlist, errr := readFileLines(options.Wordlist)
	if errr != nil {
		return errr
	}
	//extension
	if options.Extensions != "" {
		var newWordlist []string
		extensions := strings.Split(options.Extensions, ",")
		for _, ext := range extensions {
			for _, word := range wordlist {
				newWordlist = append(newWordlist, word)
				newWordlist = append(newWordlist, addExtensionToPath(word, ext))
			}
		}
		wordlist = newWordlist
	}
	c.Wordlist = wordlist

	//others
	c.Threads = options.Threads
	c.Silent = options.Silent
	c.FailureCheckTimeout = options.FailureCheckTimeout
	c.Method = options.Method
	c.FailureCounter = options.FailureConter
	c.DuplicateCounter = options.DuplicateCounter
	c.RedirectCounter = options.RedirectConter
	c.Depth = options.Depth
	c.TimeOut = options.TimeOut

	if options.RateLimit > 0 {
		c.UseRateLimit = true
		c.RateLimit = options.RateLimit
	} else {
		c.UseRateLimit = false // os this code block set this parameter false end worker do not use rate limit
		c.RateLimit = 10       //for integer divide by zero error
	}

	return nil
}
