package ignorefiles

import (
	"io/ioutil"
	"regexp"
	"strings"
)

var fileNames []string
var ignoredLineRegexp *regexp.Regexp = regexp.MustCompile("\\S*\n")

func PopulateIgnored() {
	data, err := ioutil.ReadFile(".gitignore")
	if err == nil {
		fileNames = ignoredLineRegexp.FindAllString(string(data), -1)
		fileNames = append(fileNames, ".git")
		for i, file := range fileNames {
			fileNames[i] = strings.Trim(file, "\n ")
		}
	}
}

func FileShouldBeIgnored(fileName string) bool {
	for _, file := range fileNames {
		if fileName == file {
			return true
		} else if fileName == "."+file {
			return true
		} else if fileName == "./"+file {
			return true
		}

	}

	return false
}
