package ignorefiles

import (
	"io/ioutil"
	"regexp"
	"strings"
)

var fileNames []string
var ignoredLineRegexp *regexp.Regexp = regexp.MustCompile("\\S*\n")

func PopulateIgnored(ignoreArg string) {
	data, err := ioutil.ReadFile(".gitignore")
	if err == nil {
		fileNames = ignoredLineRegexp.FindAllString(string(data), -1)
		fileNames = append(fileNames, ".git")
	}

	customArgs := strings.SplitN(ignoreArg, "=", 2)

	if len(customArgs) > 1 {
		customFiles := strings.SplitN(customArgs[1], ",", -1)
		for _, customFile := range customFiles {
			fileNames = append(fileNames, customFile)
		}
	}

	for i, file := range fileNames {
		fileNames[i] = strings.Trim(file, "\n ")
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
