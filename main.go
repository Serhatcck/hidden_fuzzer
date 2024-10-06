package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"strconv"

	"github.com/Serhatcck/hidden_fuzzer/pkg/hidden_fuzzer"
)

// Custom function to print usage message
func printUsage(flagSet *flag.FlagSet) {
	fmt.Println("Usage: hidden_fuzzer [options]")
	fmt.Println("Options:")
	flagSet.VisitAll(func(f *flag.Flag) {
		name := f.Name
		if f.DefValue != "" {
			name += fmt.Sprintf(" (default: %s)", f.DefValue)
		}
		fmt.Printf("  -%-35s %s\n", name, f.Usage)
	})
}

func main() {

	//test()
	var h bool
	var options hidden_fuzzer.Options

	flagSet := flag.NewFlagSet("Hidden Fuzzer", flag.ExitOnError) // Change to ContinueOnError
	//TO DO create new parameter for rate limit
	flagSet.BoolVar(&h, "help", false, "Show the help message")
	flagSet.StringVar(&options.Wordlist, "w", "", "Wordlist file")
	flagSet.StringVar(&options.Url, "u", "", "Target URL")
	flagSet.StringVar(&options.Extensions, "e", "", "Extensions, e.g., \".json\". Use commas for multiple extensions")
	flagSet.Var(&options.Headers, "H", "Header `\"Name: Value\"`. Use multiple -H for more headers")
	flagSet.StringVar(&options.Method, "m", "GET", "HTTP method (currently only supports GET)")
	flagSet.IntVar(&options.Threads, "t", 50, "Maximum number of threads")
	flagSet.IntVar(&options.FailureConter, "fail-counter", 3, "Number of failures to check")
	flagSet.IntVar(&options.DuplicateCounter, "dp-counter", 50, "Number of duplicates to check")
	flagSet.IntVar(&options.RedirectConter, "rd-counter", 3, "Number of redirects to check")
	flagSet.BoolVar(&options.Silent, "silent", false, "Suppress output")
	flagSet.IntVar(&options.FailureCheckTimeout, "fc-tm-out", 1, "Failure check timeout (seconds)")
	flagSet.IntVar(&options.TimeOut, "tm-out", 20, "HTTP response timeout (seconds)")
	flagSet.IntVar(&options.Depth, "depth", 3, "Subdirectory depth")
	flagSet.BoolVar(&options.ParamFuzing, "param-fuzzing", false, "For parameter fuzzing")
	flagSet.StringVar(&options.ParamValue, "param-value", "test", "Value for parameter fuzzing")
	flagSet.BoolVar(&options.Pipe, "pipe", false, "For pipe usage")
	flagSet.StringVar(&options.FilterCode, "fc", "", "Filter response HTTP status code, only one code allowed")
	flagSet.StringVar(&options.ProxyUrl, "p", "", "Proxy URL for all ongoing requests")
	flagSet.BoolVar(&options.XFFHeader, "xff", false, "Use X-F-F headers")
	flagSet.StringVar(&options.XFFValue, "xff-val", "127.0.0.1", "Value for X-F-F headers (e.g., localhost 127.0.0.1)")
	// Parse the flags

	flagSet.Usage = func() {
		printUsage(flagSet)
	}

	if err := flagSet.Parse(os.Args[1:]); err != nil {
		printUsage(flagSet)
		return
	}
	//
	if h {
		printUsage(flagSet)
		return
	}

	var conf hidden_fuzzer.Config
	err := conf.Build(options)
	if err != nil {
		log.Fatal(err.Error())
	}
	worker := hidden_fuzzer.NewWorker(&conf)
	errr := worker.Start()
	if errr != nil {
		fmt.Println("Error: " + errr.Error())
	} else {
		if !options.Pipe {
			fmt.Println("\nAnalze ended:")
			fmt.Println("")
		}

		var foundUrls []hidden_fuzzer.FoundUrl

		if options.FilterCode != "" {
			for _, resp := range worker.FoundUrls {
				if strconv.Itoa(resp.Response.StatusCode) != options.FilterCode {
					foundUrls = append(foundUrls, resp)
				}
			}
		} else {
			foundUrls = worker.FoundUrls
		}

		for _, resp := range foundUrls {
			if options.Pipe {
				fmt.Println(resp.Request.URL)
			} else {
				if resp.IsRedirect {

					//fmt.Println("RedirectedURl: " + resp.Request.URL)
					fmt.Println(resp.Request.URL + " : " + strconv.Itoa(resp.Response.StatusCode))

				} else {
					fmt.Println(resp.Request.URL + " : " + strconv.Itoa(resp.Response.StatusCode))

				}
			}

		}
	}
}

func test() {
	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()
	threadlimiter := make(chan bool, 5)
	var targets = [15]string{
		"http://localhost:8080",
		"http://localhost:8080",
		"http://localhost:8080",
		"http://localhost:8080",
		"http://localhost:8080",
		"http://localhost:8080",
		"http://localhost:8080",
		"http://localhost:8080",
		"http://localhost:8080",
		"http://localhost:8080",
		"http://localhost:8080",
		"http://localhost:8080",
		"http://localhost:8080",
		"http://localhost:8080",
		"http://localhost:8080",
	}

	for i, target := range targets {
		threadlimiter <- true

		go func(p int, t string) {
			defer func() { <-threadlimiter }()
			fmt.Println("Start", p, "----", t)
			var conf hidden_fuzzer.Config
			err := conf.Build(hidden_fuzzer.Options{
				Url:                 t,
				Wordlist:            "/Users/serhatcicek/Desktop/wordlists/demo.txt",
				Extensions:          "",
				Method:              "GET",
				Threads:             350,
				FailureConter:       2,
				DuplicateCounter:    50,
				RedirectConter:      3,
				Silent:              true,
				FailureCheckTimeout: 2,
				TimeOut:             20,
				Depth:               3,
				RateLimit:           300,
			})
			if err != nil {
				log.Fatal(err.Error())
			}
			worker := hidden_fuzzer.NewWorker(&conf)
			worker.Start()
			fmt.Println("Done", p, "----", t)
			worker.Shutdown()
		}(i, target)

	}

}
