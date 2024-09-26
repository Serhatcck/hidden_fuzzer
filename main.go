package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/Serhatcck/hidden_fuzzer/pkg/hidden_fuzzer"
)

func main() {
	var h bool
	var options hidden_fuzzer.Options

	flagSet := flag.NewFlagSet("Hidden Fuzzer", flag.ExitOnError)
	//TO DO create new parameter for rate limit
	flagSet.BoolVar(&h, "h", false, "Show the help message")
	flagSet.StringVar(&options.Url, "url", "", "Target URL")
	flagSet.StringVar(&options.Wordlist, "w", "", "Wordlist File")
	flagSet.StringVar(&options.Extensions, "e", "", "Extensions example: \".json\" for multiple usage use \",\"")
	flagSet.Var(&options.Headers, "H", "Header `\"Name: Value\"` for multiple usage use give multiple -H")
	flagSet.StringVar(&options.Method, "m", "GET", "Method (Firstly only support GET method)")
	flagSet.IntVar(&options.Threads, "t", 50, "Maximum thread number")
	flagSet.IntVar(&options.FailureConter, "fail-counter", 3, "Failure check number")
	flagSet.IntVar(&options.DuplicateCounter, "dp-counter", 50, "Duplicate counter number")
	flagSet.IntVar(&options.RedirectConter, "rd-counter", 3, "Redirect check number")
	flagSet.BoolVar(&options.Silent, "silent", false, "Silent ")
	flagSet.IntVar(&options.FailureCheckTimeout, "fc-tm-out", 1, "Failure Check Time Out (Second) ")
	flagSet.IntVar(&options.TimeOut, "tm-out", 20, "HTTP Response Time Out (Second) ")
	flagSet.IntVar(&options.Depth, "depth", 3, "Sub directory depth number")
	flagSet.BoolVar(&options.ParamFuzing, "param-fuzzing", false, "For parameter fuzzing")
	flagSet.StringVar(&options.ParamValue, "param-value", "test", "Value for parameter fuzzing")
	flagSet.BoolVar(&options.Pipe, "pipe", false, "For pipe usage")
	flagSet.StringVar(&options.FilterCode, "fc", "", "Filter response http status code, only one code status")

	flagSet.Parse(os.Args[1:])
	//
	if h {
		flagSet.Usage()
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
