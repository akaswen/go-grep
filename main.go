package main

import (
	"errors"
	"fmt"
	"go-grep/ignorefiles"
	"io/ioutil"
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

var allFiles []string
var currentOptions Options

var lineRegexp *regexp.Regexp = regexp.MustCompile(".*\n")
var errInvalidArguments = errors.New("Invalid arguments")

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func includes(arg string, slice []string) bool {
	for _, str := range slice {
		if str == arg {
			return true
		}
	}

	return false
}

func setOptions(o *Options) {
	if len(os.Args) < 2 {
		check(errInvalidArguments)
	}

	o.UserRegexp = regexp.MustCompile(os.Args[1])
	o.ShouldIgnoreFiles = includes("-I", os.Args)
}

func printFileMatches(fileName string, wg *sync.WaitGroup) {
	defer wg.Done()

	dat, err := ioutil.ReadFile(fileName)
	check(err)
	lines := lineRegexp.FindAllString(string(dat), -1)
	for i, line := range lines {
		if currentOptions.UserRegexp.MatchString(line) {
			match := LineMatch{fileName, line, i + 1}
			fmt.Printf("%v - %v: %v\n", match.FileName, match.LineNumber, match.Line)
		}
	}
}

func getMatches(path string, wg *sync.WaitGroup) {
	files, err := ioutil.ReadDir(path)
	check(err)
	for _, file := range files {
		fileName := path + "/" + file.Name()
		if currentOptions.ShouldIgnoreFiles && ignorefiles.FileShouldBeIgnored(fileName) {
			continue
		}

		if file.IsDir() {
			getMatches(fileName, wg)
		} else {
			wg.Add(1)
			go printFileMatches(fileName, wg)
		}
	}
}

func getAllFiles(path string) {
	files, err := ioutil.ReadDir(path)
	check(err)
	for _, file := range files {
		fileName := path + "/" + file.Name()
		if ignorefiles.FileShouldBeIgnored(fileName) {
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

	setOptions(&currentOptions)

	if currentOptions.ShouldIgnoreFiles {
		ignorefiles.PopulateIgnored()
	}

	getAllFiles(".")

	var wg sync.WaitGroup
	getMatches(".", &wg)
	wg.Wait()

	elapsed := time.Since(start)
	fmt.Println("search took: ", elapsed)
}
