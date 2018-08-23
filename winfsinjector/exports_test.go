package winfsinjector

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

func SetReadDir(f func(string) ([]os.FileInfo, error)) {
	readDir = f
}

func ResetReadDir() {
	readDir = ioutil.ReadDir
}
