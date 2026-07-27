package pkg

import "os"

func MustReadFile(filePath string) string {
	f, err := os.ReadFile(filePath)
	if err != nil {
		panic(err)
	}
	return string(f)
}
