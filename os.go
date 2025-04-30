package migration

import (
	"fmt"
	"github.com/sandwich-go/boost/xos"
	"os"
)

var (
	CwdDir = "/usr/local/bin"
)

func init() {
	if xos.IsGoRun() {
		CwdDir = "/tmp"
	} else {
		if cwd, err := os.Getwd(); err == nil {
			CwdDir = cwd
		}
	}
}

func Chdir(dest string, logger *Logger) (deferFunc func(), err error) {
	deferFunc = func() {}
	var cwdDir string
	if cwdDir, err = os.Getwd(); err == nil {
		deferFunc = func() {
			logger.Info(fmt.Sprintf("Chdir from %s back to %s", dest, cwdDir))
			_ = os.Chdir(cwdDir)
		}
	} else if CwdDir != "" {
		cwdDir = CwdDir
		deferFunc = func() {
			logger.Info(fmt.Sprintf("Chdir from %s back to %s", dest, CwdDir))
			_ = os.Chdir(CwdDir)
		}
	}
	logger.Info(fmt.Sprintf("Chdir from %s to %s", cwdDir, dest))
	if err = os.Chdir(dest); err != nil {
		return deferFunc, fmt.Errorf("got err: %s while chdir: %s", err.Error(), dest)
	}

	return deferFunc, nil
}
