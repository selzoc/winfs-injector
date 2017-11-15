package injector

import "io/ioutil"

func SetReadFile(f func(string) ([]byte, error)) {
	readFile = f
}

func ResetReadFile() {
	readFile = ioutil.ReadFile
}
