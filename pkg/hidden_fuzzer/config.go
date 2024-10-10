package hidden_fuzzer

import (
	"context"
	"errors"
	"net/url"
	"strings"
)

type Config struct {
	Context                 context.Context
	Target                  string
	Url                     *url.URL
	Silent                  bool
	Wordlist                []string
	Extensions              []string
	Threads                 int
	Headers                 map[string]string
	FailureCheckTimeout     int
	TimeOut                 int
	Method                  string
	FailureCounter          int
	DuplicateCounter        int
	RedirectCounter         int
	Depth                   int
	RateLimit               int
	UseRateLimit            bool
	ParamFuzing             bool
	ParamValue              string
	ProxyUrl                string
	MaxBodyLengthForCompare int64
}

func (c *Config) Build(options Options) error {
	c.Headers = make(map[string]string)

	//headers

	for key, value := range options.Headers {
		c.Headers[key] = value
	}
	if c.Headers["User-Agent"] == "" {
		c.Headers["User-Agent"] = "User-Agent Mozilla/5.0 (Macintosh; Intel Mac OS X 10.15; rv:130.0) Gecko/20100101 Firefox/130.0"
	}
	if options.XFFHeader {
		// getXFFHeaders fonksiyonunu çağırın ve dönen haritayı ekleyin
		for k, v := range getXFFHeaders(options.XFFValue) {
			c.Headers[k] = v // Yeni anahtar-değer çiftlerini ekle
		}
	}

	//url
	url, err := getURL(options.Url)
	if err {
		return errors.New("given url isn't url")
	}
	c.Url = url

	//wordlist

	var wordlist []string
	if len(options.WordlistStringArray) > 0 {
		wordlist = options.WordlistStringArray
	} else {
		var errr error
		wordlist, errr = readFileLines(options.Wordlist)
		if errr != nil {
			return errr
		}
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

	c.ParamFuzing = options.ParamFuzing
	c.ParamValue = options.ParamValue
	c.ProxyUrl = options.ProxyUrl
	c.MaxBodyLengthForCompare = options.MaxBodyLengthForCompare

	//if param fuzzing is true do not handle 403 or directories.
	//do not process sub directory depth
	if c.ParamFuzing {
		c.Depth = 0
	}

	if options.Pipe {
		c.Silent = true
	}

	if options.RateLimit > 0 {
		c.UseRateLimit = true
		c.RateLimit = options.RateLimit
	} else {
		c.UseRateLimit = false // os this code block set this parameter false end worker do not use rate limit
		c.RateLimit = 10       //for integer divide by zero error
	}

	return nil
}

func (c *Config) SetUrl(urlstring string) error {
	// url
	url, err := getURL(urlstring)
	if err {
		return errors.New("given url isn't url")
	}
	c.Url = url
	return nil
}
