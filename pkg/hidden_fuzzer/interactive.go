package hidden_fuzzer

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"
)

type interactive struct {
	Worker *Worker
	paused bool
}

func HandleInteractive(worker *Worker) error {
	i := interactive{worker, false}

	inreader := bufio.NewScanner(os.Stdin)
	inreader.Split(bufio.ScanLines)

	for inreader.Scan() {
		if worker.isrunning {
			i.handleInput(inreader.Bytes())
		} else {
			return nil
		}
	}
	return nil
}

func (i *interactive) handleInput(in []byte) {
	instr := string(in)
	args := strings.Split(strings.TrimSpace(instr), " ")
	if len(args) == 1 && args[0] == "" {
		// Enter pressed - toggle interactive state
		i.paused = !i.paused
		if i.paused {
			i.Worker.sleep()
			i.Worker.isInteractiveWindow = true
			time.Sleep(500 * time.Millisecond)
			i.printBanner()
		} else {
			i.Worker.sleep()
		}
	} else {
		switch args[0] {
		case "?":
			i.printHelp()
		case "help":
			i.printHelp()
		case "resume":
			i.paused = false
			i.Worker.isInteractiveWindow = false
			i.Worker.resume()
		case "queueskip":
			i.Worker.WorkQueue = nil
			i.Worker.skipQueue = true
			fmt.Printf("Queue Count: %d\n", len(i.Worker.WorkQueue))
		case "current":
			fmt.Printf("Current Target: %s\n", i.Worker.currentTarget)
		default:
			if i.paused {
				fmt.Printf("Unknown command: \"%s\". Enter \"help\" for a list of available commands", args[0])
			} else {
				fmt.Printf("Nope")
			}
		}
	}

	if i.paused {
		i.printPrompt()
	}
}

func (i *interactive) printBanner() {
	fmt.Printf("entering interactive mode\ntype \"help\" for a list of commands, or ENTER to resume.\n")
}

func (i *interactive) printPrompt() {
	fmt.Printf("> ")
}

func (i *interactive) printHelp() {
	help := `
available commands:
 queueskip                - advance to the next queued job
 restart                  - restart and resume the current ffuf job
 resume                   - resume current ffuf job (or: ENTER) 
 current				  - show current queue target
`
	fmt.Print(help)
}
