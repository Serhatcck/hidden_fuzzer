# URL Fuzzing Tool

This project is a URL fuzzing tool that helps to discover hidden paths and resources on a target web application. It supports multithreading, custom HTTP headers, and customizable request parameters to optimize fuzzing performance.

The tool uses the similarity algorithm from [hyperjumptech/beda](https://github.com/hyperjumptech/beda) to evaluate if the results are valid or false positives during the fuzzing process.

## Features

- URL fuzzing with custom wordlists
- Customizable request method (currently supports only `GET`)
- Ability to add custom headers
- Multi-threading support for faster fuzzing
- Detects potential false positives using a similarity algorithm
- Silent mode for quiet output
- Custom timeouts for failure and response handling
- Adjustable depth for fuzzing subdirectories

## Installation

To install and use the tool as a module, you can include it in your project from the following repository:

```bash
go get github.com/Serhatcck/hidden_fuzzer
```

## Usage
The tool can be run from the command line with the following options:
```bash
Usage: hidden_fuzzer [options]

Options:
  -h                     Show the help message
  -url string            Target URL
  -w string              Wordlist file
  -e string              Extensions (e.g., ".json"). For multiple extensions, use commas.
  -H "Name: Value"       Custom headers. For multiple headers, pass multiple `-H` options.
  -m string              Request method (currently only supports GET)
  -t int                 Maximum number of threads (default: 50)
  -fail-counter int      Number of allowed failures before stopping (default: 3)
  -dp-counter int        Number of duplicate responses before stopping (default: 50)
  -rd-counter int        Number of allowed redirects before stopping (default: 3)
  -silent                Run in silent mode
  -fc-tm-out int         Failure check timeout in seconds (default: 1)
  -tm-out int            HTTP response timeout in seconds (default: 20)
  -depth int             Maximum subdirectory depth to fuzz (default: 3)
```


## Example
Fuzzing a target URL with a custom wordlist and headers:
```bash
hidden_fuzzer -url http://example.com -w wordlist.txt -H "Authorization: Bearer token" -e ".php,.html" -t 100
```
This command fuzzes the target http://example.com using the specified wordlist and headers, searching for .php and .html files with 100 threads.

## Module Usage

You can use this tool as a module in your Go project. Import it as follows:
```go
import "github.com/Serhatcck/hidden_fuzzer"

```


### Acknowledgments

Special thanks to the following repositories for their inspiration and contributions:

- [ffuf](https://github.com/ffuf/ffuf) for providing inspiration on how to structure and optimize the fuzzing process.
- [hyperjumptech/beda](https://github.com/hyperjumptech/beda) for the similarity algorithm used in detecting false positives.