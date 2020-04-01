package main

import (
	"crypto/sha256"
	"io"
	"log"
	"os"
)

const invalidHashString = "invalid"

func Sha256Sum(fileName string) (string, error) {
	f, err := os.Open(fileName)
	if err != nil {
		return invalidHashString, err
	}
	defer func() {
		err = f.Close()
		if err != nil {
			log.Println("Cannot close file " + fileName)
		}
	}()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return invalidHashString, err
	}

	return string(h.Sum(nil)), nil
}
