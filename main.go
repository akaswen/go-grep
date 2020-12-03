package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"time"

	"go-grep/pkg/search"
)

type options struct {
	useAck bool
}

func parseArgs() (opts options, err error) {
	flag.BoolVar(&opts.useAck, "a", false, "uses ack instead")
	flag.Parse()

	return opts, nil
}

func useAck(searchTerm, filePath string) {
	fmt.Println(searchTerm, filePath)
	cmd := exec.Command("ack", searchTerm, filePath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Run()
}

func getArgs() (searchTerm, filePath string, err error) {
	searchTerm = flag.Arg(0)
	filePath = flag.Arg(1)

	if searchTerm == "" || filePath == "" {
		return "", "", fmt.Errorf("need at least 2 arguments")
	}

	return
}

func main() {
	start := time.Now()

	opts, err := parseArgs()
	searchTerm, filePath, err := getArgs()
	if err != nil {
		fmt.Println(err)
		return
	}

	if err != nil {
		panic(err)
	}

	if opts.useAck {
		useAck(searchTerm, filePath)
	} else {
		search.Search(searchTerm, filePath)
	}

	fmt.Println("Search took: ", time.Since(start))
}
