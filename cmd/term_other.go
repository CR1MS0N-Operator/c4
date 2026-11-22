//go:build !linux

package cmd

import (
	"errors"
	"os"
)

func isTerminal(f *os.File) bool {
	fi, err := f.Stat()
	return err == nil && fi.Mode()&os.ModeCharDevice != 0
}

func readHidden(*os.File, string) (string, error) {
	return "", errors.New("hidden terminal input is not supported on this platform; pipe the value on stdin")
}
