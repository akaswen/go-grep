package search

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"regexp"
	"sync"
	"syscall"

	"github.com/as/hue"
)

type result struct {
	filePath   string
	lineNumber int
	line       string
}

var searchRegex *regexp.Regexp
var printResultWriter *hue.RegexpWriter
var fileColor = hue.Red
var matchColor = hue.Green
var numbersColor = hue.Blue
var logger = log.New(os.Stderr, "", log.Lmicroseconds)

func logErr(err error) {
	logger.Println(err)
}

func readPwd(path string) ([]os.FileInfo, error) {
	file, err := os.Open(path)
	if err != nil {
		logErr(err)
		return make([]os.FileInfo, 0), err
	}

	defer file.Close()

	files, err := file.Readdir(-1)
	if err != nil {
		logErr(err)
		return make([]os.FileInfo, 0), err
	}

	return files, nil
}

func sanitize(filePath string) string {
	if filePath[len(filePath)-1] == '/' {
		return filePath
	} else {
		return filePath + "/"
	}
}

func printResults(results chan result, ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case r := <-results:
			output := fmt.Sprintf("%v - %v: %v\n", r.filePath, r.lineNumber, r.line)
			printResultWriter.WriteString(output)
		}
	}
}

func printRestOfResults(results chan result) {
	for {
		select {
		case r := <-results:
			output := fmt.Sprintf("%v - %v: %v\n", r.filePath, r.lineNumber, r.line)
			printResultWriter.WriteString(output)
		default:
			return
		}
	}
}

type WaitGroup struct {
	mu    sync.Mutex
	wg    sync.WaitGroup
	count int
}

func (wg *WaitGroup) Add() {
	wg.mu.Lock()
	defer wg.mu.Unlock()
	wg.wg.Add(1)
	wg.count += 1
}

func (wg *WaitGroup) Done() {
	wg.mu.Lock()
	defer wg.mu.Unlock()
	wg.wg.Done()
	wg.count -= 1
}

func Search(searchTerm, filePath string) {
	searchRegex = regexp.MustCompile(fmt.Sprintf("(%v)", searchTerm))

	printResultWriter = hue.NewRegexpWriter(os.Stdout)
	match := hue.New(matchColor, hue.Default)
	file := hue.New(fileColor, hue.Default)
	number := hue.New(numbersColor, hue.Default)
	printResultWriter.AddRuleString(match, searchTerm)
	printResultWriter.AddRuleString(file, `^\S+`)
	printResultWriter.AddRuleString(number, `\d+:`)

	results := make(chan result, 1000)
	files := make(chan string, 1000)
	var wg sync.WaitGroup
	ctx, cancel := context.WithCancel(context.Background())

	go printResults(results, ctx)

	var rLimit syscall.Rlimit
	err := syscall.Getrlimit(syscall.RLIMIT_NOFILE, &rLimit)
	if err != nil {
		panic(err)
	}

	for i := uint64(0); i < rLimit.Cur - 10; i++ {
		go processFiles(files, results, &wg)
	}

	searchDirectories(filePath, files, &wg)
	wg.Wait()
	cancel()
	printRestOfResults(results)
}

func processFiles(files chan string, results chan result, wg *sync.WaitGroup) {
	for {
		filePath := <-files
		searchFile(filePath, results)
		wg.Done()
	}
}

func searchDirectories(filePath string, files chan string, wg *sync.WaitGroup) {
	basePath := sanitize(filePath)
	fileInfos, err := readPwd(basePath)
	if err != nil {
		logErr(err)
		return
	}

	for _, f := range fileInfos {
		if f.IsDir() && f.Name() != ".git" {
			searchDirectories(basePath+f.Name(), files, wg)
		} else {
			if f.Mode().IsRegular() {
				wg.Add(1)
				files <- basePath + f.Name()
			}
		}
	}
}

func searchFile(filePath string, ch chan result) {
	file, err := os.Open(filePath)
	if err != nil {
		logErr(err)
		return
	}

	defer file.Close()

	scanner := bufio.NewScanner(file)
	lineNumber := 1
	for scanner.Scan() {
		line := scanner.Text()
		if searchRegex.MatchString(line) {
			ch <- result{
				filePath:   filePath,
				lineNumber: lineNumber,
				line:       line,
			}
		}
		lineNumber += 1
	}
}
