package migration

import (
	"fmt"
	"os"
)

func Chdir(dest string) (deferFunc func(), err error) {
	deferFunc = func() {}
	if cwdDir, err := os.Getwd(); err == nil {
		deferFunc = func() {
			_ = os.Chdir(cwdDir)
		}
	} else {
		return deferFunc, fmt.Errorf("got err: %s get current dir", err.Error())
	}

	if err = os.Chdir(dest); err != nil {
		return deferFunc, fmt.Errorf("got err: %s while chdir: %s", err.Error(), dest)
	}

	return deferFunc, nil
}
