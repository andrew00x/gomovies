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

func ReadLines(file string, lineHandler func(string) bool) error {
	f, err := os.Open(file)
	if err != nil {
		return err
	}
	defer f.Close()
	input := bufio.NewScanner(f)
	for input.Scan() {
		line := input.Text()
		if !lineHandler(line) {
			break
		}
	}
	return nil
}
