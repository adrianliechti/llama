package template

import (
	"os"
)

func include(path string) string {
	data, err := os.ReadFile(path)

	if err != nil {
		panic(err)
	}

	return string(data)
}
