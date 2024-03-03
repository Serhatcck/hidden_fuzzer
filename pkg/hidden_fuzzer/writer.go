package hidden_fuzzer

import (
	"fmt"
	"os"
)

const (
	Clear_Terminal = "\r\033[K"
)

func WriteStatus(worker *Worker, stat int, counter int) {
	if !worker.Config.Silent {
		fmt.Fprintf(os.Stderr, "%s[Target: %s] -- [Req per second: %d] -- [Status: %d/%d]", Clear_Terminal, worker.currentTarget, stat, counter, len(worker.WorkQueue))
	}
}

func WriteFound(worker *Worker, url string) {
	if !worker.Config.Silent {
		fmt.Fprintf(os.Stderr, "%s%s\n", Clear_Terminal, url)
	}
}

func WriteFailure(worker *Worker, data string) {
	if !worker.Config.Silent {
		fmt.Fprintf(os.Stderr, "%s %d. %s\n", Clear_Terminal, worker.Config.FailureCounter, data)
	}
}
