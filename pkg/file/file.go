package file

import (
	"bufio"
	"os"
)

func Exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return true, err
}

func ReadLines(file string, lineHandler func(string) bool) (err error) {
	var f *os.File
	if f, err = os.Open(file); err != nil {
		return
	}
	defer func() {
		if clsErr := f.Close(); clsErr != nil {
			err = clsErr
		}
	}()
	input := bufio.NewScanner(f)
	for input.Scan() {
		line := input.Text()
		if !lineHandler(line) {
			break
		}
	}
	return
}
