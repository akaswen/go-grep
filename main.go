package main

import (
	"errors"
	"fmt"
	"go-grep/ignorefiles"
	"os"
	"regexp"
	"sync"
	"time"
)

type LineMatch struct {
	FileName   string
	Line       string
	LineNumber int
}

type Options struct {
	ShouldIgnoreFiles bool
	UserRegexp        *regexp.Regexp
}

var currentOptions Options
var lineRegexp *regexp.Regexp = regexp.MustCompile(".*\n")
var errInvalidArguments = errors.New("Invalid arguments")

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func findArg(arg string, slice []string) (string, bool) {
	for _, str := range slice {
		if str[:2] == arg {
			return str, true
		}
	}

	return "", false
}

func setOptions(o *Options) {
	if len(os.Args) < 2 {
		check(errInvalidArguments)
	}

	o.UserRegexp = regexp.MustCompile(os.Args[1])
	var ignoreArg string
	ignoreArg, o.ShouldIgnoreFiles = findArg("-I", os.Args)
	if o.ShouldIgnoreFiles {
		ignorefiles.PopulateIgnored(ignoreArg)
	}
}

func getFileData(fileName string) []byte {
	file, err := os.Open(fileName)
	check(err)
	info, err := file.Stat()
	check(err)
	data := make([]byte, info.Size())
	_, err = file.Read(data)
	check(err)
	file.Close()

	return data
}

func printFileMatches(fileName string, wg *sync.WaitGroup) {
	defer wg.Done()

	data := getFileData(fileName)

	lines := lineRegexp.FindAllString(string(data), -1)
	for i, line := range lines {
		if currentOptions.UserRegexp.MatchString(line) {
			match := LineMatch{fileName, line, i + 1}
			fmt.Printf("%v - %v: %v\n", match.FileName, match.LineNumber, match.Line)
		}
	}

}

func exploreFiles(path string, wg *sync.WaitGroup) {
	file, err := os.Open(path)
	check(err)
	files, err := file.Readdir(-1)
	check(err)
	file.Close()

	for _, file := range files {
		fileName := path + "/" + file.Name()
		if currentOptions.ShouldIgnoreFiles && ignorefiles.FileShouldBeIgnored(fileName) {
			continue
		}

		if file.IsDir() {
			exploreFiles(fileName, wg)
		} else {
			wg.Add(1)
			go printFileMatches(fileName, wg)
		}
	}
}

func main() {
	start := time.Now()

	setOptions(&currentOptions)

	var wg sync.WaitGroup
	exploreFiles(".", &wg)
	wg.Wait()

	elapsed := time.Since(start)
	fmt.Println("search took: ", elapsed)
}
