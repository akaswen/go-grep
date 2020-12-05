package search

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"sync"
	"time"

	"github.com/as/hue"
)

type result struct {
	filePath   string
	lineNumber int
	line       string
}

var searchRegex *regexp.Regexp
var printResultWriter *hue.RegexpWriter
var processFilesThreads int = 300
var fileColor = hue.Red
var matchColor = hue.Green
var numbersColor = hue.Blue

func check(err error) {
	if err != nil {
		panic(err)
	}
}

func readPwd(path string) ([]os.FileInfo, error) {
	file, err := os.Open(path)
	if err != nil {
		return make([]os.FileInfo, 0), err
	}

	defer file.Close()

	files, err := file.Readdir(-1)
	check(err)

	return files, nil
}

func sanitize(filePath string) string {
	if filePath[len(filePath)-1] == '/' {
		return filePath
	} else {
		return filePath + "/"
	}
}

func printResults(results chan result) {
	for {
		r := <-results
		output := fmt.Sprintf("%v - %v: %v\n", r.filePath, r.lineNumber, r.line)
		printResultWriter.WriteString(output)
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

	results := make(chan result, 16000)
	files := make(chan string, 16000)
	var wg WaitGroup

	go printResults(results)

	for i := 0; i < processFilesThreads; i++ {
		go processFiles(files, results, &wg)
	}

	searchDirectories(filePath, files, &wg)
	wg.wg.Wait()
	time.Sleep(100 * time.Millisecond)
}

func processFiles(files chan string, results chan result, wg *WaitGroup) {
	for {
		filePath := <-files
		searchFile(filePath, results)
		wg.Done()
	}
}

func searchDirectories(filePath string, files chan string, wg *WaitGroup) {
	basePath := sanitize(filePath)
	fileInfos, err := readPwd(basePath)
	if err != nil {
		if match, _ := regexp.MatchString("operation not permitted", err.Error()); match {
			return
		} else if match, _ := regexp.MatchString("permission denied", err.Error()); match {
			return
		} else {
			fmt.Println("the error: ", err.Error())
			panic(err)
		}
	}

	for _, f := range fileInfos {
		if f.IsDir() && f.Name() != ".git" {
			searchDirectories(basePath+f.Name(), files, wg)
		} else {
			if f.Mode().IsRegular() {
				wg.Add()
				files <- basePath + f.Name()
			}
		}
	}
}

func searchFile(filePath string, ch chan result) {
	file, err := os.Open(filePath)
	if err != nil {
		if match, _ := regexp.MatchString("operation not permitted", err.Error()); match {
			return
		} else if match, _ := regexp.MatchString("permission denied", err.Error()); match {
			return
		} else {
			panic(err)
		}
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
