package migration

import (
	"fmt"
	"os"
)

func Chdir(dest string, logger *Logger) (deferFunc func(), err error) {
	deferFunc = func() {}
	var cwdDir string
	if cwdDir, err = os.Getwd(); err == nil {
		deferFunc = func() {
			logger.Info(fmt.Sprintf("Chdir from %s back to %s", dest, cwdDir))
			_ = os.Chdir(cwdDir)
		}
	} else {
		return deferFunc, fmt.Errorf("got err: %s get current dir", err.Error())
	}
	logger.Info(fmt.Sprintf("Chdir from %s to %s", cwdDir, dest))
	if err = os.Chdir(dest); err != nil {
		return deferFunc, fmt.Errorf("got err: %s while chdir: %s", err.Error(), dest)
	}

	return deferFunc, nil
}
