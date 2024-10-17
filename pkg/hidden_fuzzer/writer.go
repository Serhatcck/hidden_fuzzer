package hidden_fuzzer

import (
	"fmt"
	"os"
	"strconv"
)

const (
	Clear_Terminal = "\r\033[K"
)

func WriteStatus(worker *Worker, stat int, counter int) {
	if !worker.Config.Silent {
		fmt.Fprintf(os.Stderr, "%s[Target: %s] -- [Req per second: %d] -- [Status: %d/%d]", Clear_Terminal, worker.currentTarget, stat, counter, len(worker.WorkQueue))
	}
}

func WriteFound(worker *Worker, queue WorkQueue, response Response) {
	if !worker.Config.Silent {
		var print = "[" + queue.Req.Method + "] "
		if queue.RedirectConter > 0 {
			for _, rsp := range queue.RedirectQueue {
				print += "[" + rsp.Request.URL + "] -> "
			}
		}
		print += response.URL + " : " + strconv.Itoa(response.StatusCode)

		fmt.Fprintf(os.Stderr, "%s%s\n", Clear_Terminal, print)
	}
}

func WriteStr(worker *Worker, str string) {
	if !worker.Config.Silent {
		fmt.Fprintf(os.Stderr, "%s%s\n", Clear_Terminal, str)
	}
}

func WriteFailure(worker *Worker, data string) {
	if !worker.Config.Silent {
		fmt.Fprintf(os.Stderr, "%s %d. %s\n", Clear_Terminal, worker.Config.FailureCounter, data)
	}
}
