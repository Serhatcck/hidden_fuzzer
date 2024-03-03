# Hidden Fuzzer

Hidden Fuzzer is a tool for web application testing that allows you to fuzz URLs with various parameters.

## Usage

### Installation

You can install Hidden Fuzzer by cloning the repository and building the binary:

```bash
git clone https://github.com/yourusername/hidden_fuzzer.git
cd hidden_fuzzer
go build -o hidden_fuzzer main.go
```

## Command-line Parameters
- `-h`: Show the help message.
- `-url`: Specify the target URL.
- `-w`: Specify the wordlist file.
- `-e`: Specify extensions. Example: ".json" For multiple extensions, use comma-separated values.
- `-H`: Specify HTTP headers in the format "Name: Value". For multiple headers, use multiple `-H` flags.
- `-m`: Specify the HTTP method. (Currently only supports GET method)
- `-t`: Specify the maximum number of threads.
- `-fail-counter`: Specify the number of failure checks.
- `-dp-counter`: Specify the number of duplicate counters.
- `-rd-counter`: Specify the number of redirect checks.
- `-silent`: Enable silent mode.
- `-tm-out`: Specify the failure check time-out in seconds.
- `-depth`: Specify the sub-directory depth number.

## Example
This example command performs URL fuzzing on the target `https://example.com` using the wordlist file `wordlist.txt`, with extensions `.php` and `.html`, sets a custom `User-Agent` header.

```bash
hidden_fuzzer -url https://example.com -w wordlist.txt -e .php,.html -H "User-Agent: Mozilla/5.0" 
```