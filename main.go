package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"regexp"
	"os"
	"strings"
	"sync"
	"time"
)

type LineMatch struct {
	FileName string
	Line string
	LineNumber int
}

var ignoredFiles []string
var allFiles []string

var lineRegexp *regexp.Regexp = regexp.MustCompile(".*\n")
var userRegexp *regexp.Regexp
var errInvalidArguments = errors.New("Invalid arguments")

func check(e error) {
    if e != nil {
        panic(e)
    }
}

func fileShouldBeIgnored(fileName string) bool {
	ignoredFiles = append(ignoredFiles, ".git")
	for _, file := range ignoredFiles {
		if fileName == file {
			return true
		} else if fileName == "." + file {
			return true
		} else if fileName == "./" + file {
			return true
		}

	}

	return false
}

func printMatches(fileName string, wg *sync.WaitGroup) {
	defer wg.Done()

	f, err := os.Open(fileName)
	check(err)
	info, err := f.Stat()
	check(err)
	size := info.Size()
	dat := make([]byte, size)
	_, err = f.Read(dat)
	check(err)
	f.Close()
	lines := lineRegexp.FindAllString(string(dat), -1)
	for i, line := range lines {
		if userRegexp.MatchString(line) {
			match := LineMatch{fileName, line, i + 1}
			fmt.Printf("%v - %v: %v\n", match.FileName, match.LineNumber, match.Line)
		}
	}
}

func ignoreFiles() {
	data, err := ioutil.ReadFile(".gitignore")
	if err == nil {
		ignoredLineRegexp := regexp.MustCompile("\\S*\n")
		ignoredFiles = ignoredLineRegexp.FindAllString(string(data), -1)
		for i, file := range ignoredFiles {
			ignoredFiles[i] = strings.Trim(file, "\n ")
		}
	}
}

func getAllFiles(path string) {
	files, err := ioutil.ReadDir(path)
	check(err)
	for _, file := range files {
		fileName := path + "/" + file.Name()
		if fileShouldBeIgnored(fileName) {
			continue
		}
		if file.IsDir() {
			getAllFiles(fileName)
		} else {
			allFiles = append(allFiles, fileName)
		}
	}
}

func main() {
	start := time.Now()

	if len(os.Args) != 2 {
		check(errInvalidArguments)
	}
	userRegexp = regexp.MustCompile(os.Args[1])

	ignoreFiles()
	getAllFiles(".")

	var wg sync.WaitGroup

	for _, fileName := range allFiles {
		wg.Add(1)
		go printMatches(fileName, &wg)
	}

	wg.Wait()
	elapsed := time.Since(start)
	fmt.Println("search took: ", elapsed)
}
