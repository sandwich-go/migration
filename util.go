package migration

import (
	"github.com/sandwich-go/boost/xos"
	"os"
)

var FileGetContents = xos.FileGetContents
var FilePutContents = xos.FilePutContents

func fileExists(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}
	return false
}
