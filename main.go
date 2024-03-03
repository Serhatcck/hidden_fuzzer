package main

import (
	"flag"
	"fmt"
	"hidden_fuzzer/pkg/hidden_fuzzer"
	"log"
	"os"
)

func main() {
	var h bool
	var options hidden_fuzzer.Options

	flagSet := flag.NewFlagSet("Hidden Fuzzer", flag.ExitOnError)
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
	flagSet.IntVar(&options.Timeout, "tm-out", 1, "Failure Check Time Out Second ")
	flagSet.IntVar(&options.Depth, "depth", 3, "Sub directory depth number")

	flagSet.Parse(os.Args[1:])

	if h {
		flagSet.Usage()
		return
	}
	/*conf := hidden_fuzzer.Config{
		//Url: "https://trendyol.com",
		//Url: "http://testphp.vulnweb.com",
		Target: "http://192.168.1.106",
		//Wordlist: "/usr/share/wordlist/SecLists/Discovery/Web-Content/raft-small-directories-lowercase.txt",
		Wordlist: "wordlist.txt",
		Headers: map[string]string{
			"User-Agent": "Random",
		},
		Method:           "GET",
		Threads:          50,
		FailureCounter:   3,
		DuplicateCounter: 5,
		RedirectCounter:  3,
	}*/
	var conf hidden_fuzzer.Config
	err := conf.Build(options)
	if err != nil {
		log.Fatal(err.Error())
	}
	worker := hidden_fuzzer.NewWorker(&conf)
	worker.Start()

	fmt.Println("Analze ended:")
	fmt.Println("")

	for _, resp := range worker.FoundUrls {
		if resp.IsRedirect {

			//fmt.Println("RedirectedURl: " + resp.Request.URL)
			fmt.Println(resp.Request.URL)

		} else {
			fmt.Println(resp.Request.URL)

		}
	}
	/*var conf hidden_fuzzer.Config
	err := conf.Build(options)
	if err != nil {
		log.Fatal(err.Error())
	}
	os.Exit(0)
	wordlist, err := readFileLines(conf.Wordlist)
	if err != nil {
		fmt.Println(err)
	}
	conf.Wordlists = wordlist
	worker := hidden_fuzzer.NewWorker(&conf)
	worker.Start()

	fmt.Println("Analze ended:")
	fmt.Println("")

	for _, resp := range worker.FoundUrls {
		if resp.IsRedirect {

			fmt.Println("RedirectedURl: " + resp.Request.URL)
		} else {
			fmt.Println(resp.Request.URL)

		}
	}*/

}
