package injector

import (
	"io/ioutil"
	"os"
)

func SetReadFile(f func(string) ([]byte, error)) {
	readFile = f
}

func ResetReadFile() {
	readFile = ioutil.ReadFile
}

func SetRemoveAll(f func(string) error) {
	removeAll = f
}

func ResetRemoveAll() {
	removeAll = os.RemoveAll
}
