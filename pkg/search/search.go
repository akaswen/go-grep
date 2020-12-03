package search

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"regexp"
	"sync"
)

type result struct {
	filePath   string
	lineNumber int
	line       string
}

var searchRegex *regexp.Regexp

func check(err error) {
	if err != nil {
		panic(err)
	}
}

func readPwd(path string) []os.FileInfo {
	file, err := os.Open(path)
	check(err)
	defer file.Close()

	files, err := file.Readdir(-1)
	check(err)

	return files
}

func sanitize(filePath string) string {
	if filePath[len(filePath)-1] == '/' {
		return filePath
	} else {
		return filePath + "/"
	}
}

func Search(searchTerm, filePath string) {
	searchRegex = regexp.MustCompile(searchTerm)

	ch := make(chan result, 1)
	ctx, cancel := context.WithCancel(context.Background())

	go func(ch chan result, ctx context.Context) {
		for {
			select {
			case <-ctx.Done():
				break
			case result := <-ch:
				fmt.Printf("%v - %v: %v\n", result.filePath, result.lineNumber, result.line)
			default:
				continue
			}
		}
	}(ch, ctx)

	var wg sync.WaitGroup

	wg.Add(1)
	go exploreFiles(filePath, ch, &wg)
	wg.Wait()
	cancel()
}

func exploreFiles(filePath string, ch chan result, wg *sync.WaitGroup) {
	basePath := sanitize(filePath)
	files := readPwd(basePath)
	for _, f := range files {
		if f.IsDir() {
			wg.Add(1)
			go exploreFiles(basePath+f.Name(), ch, wg)
		} else {
			searchFile(basePath+f.Name(), ch)
		}
	}
	wg.Done()
}

func searchFile(filePath string, ch chan result) {
	file, err := os.Open(filePath)
	if err != nil {
		check(err)
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
