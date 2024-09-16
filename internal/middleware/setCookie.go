package middleware

import (
	"bufio"
	"os"
)

const secretKeyPath = "tmp/key.txt"

var secretKey, _ = getKey(secretKeyPath)

func getKey(filepath string) (string, error) {
	file, err := os.OpenFile(filepath, os.O_RDONLY, 0666)
	if err != nil {
		return "", err
	}
	defer file.Close()

	var line string

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line = scanner.Text()
	}
	if err = scanner.Err(); err != nil {
		return "", err
	}
	return line, nil
}
