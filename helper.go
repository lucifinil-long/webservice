package webservice

import (
	"os"
	"strings"
)

// getCurrentDirectory returns current program location path
func getCurrentDirectory() string {
	dir, err := os.Getwd()
	if err != nil {
		dir = ""
	}
	return strings.Replace(dir, "\\", "/", -1)
}

// ifDirExists test whether specified path is a existed dir
func ifDirExists(path string) bool {
	s, err := os.Stat(path)
	if err != nil {
		if os.IsExist(err) {
			return s.IsDir()
		}
		return false
	}
	return s.IsDir()
}
