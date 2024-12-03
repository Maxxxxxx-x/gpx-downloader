package utils

import (
	"crypto/sha512"
	"crypto/subtle"
	"encoding/hex"
	"io"
	"os"
)

func GenerateFileHash(file *os.File) (string, error) {
    hasher := sha512.New()

    if _, err := file.Seek(0, 0); err != nil {
        return "", err
    }

	if _, err := io.Copy(hasher, file); err != nil {
		return "", err
	}

	return hex.EncodeToString(hasher.Sum(nil)), nil
}

func CompareFileAndHash(file *os.File, hash string) (bool, error) {
	fileHash, err := GenerateFileHash(file)
	if err != nil {
		return false, err
	}

	return subtle.ConstantTimeCompare([]byte(fileHash), []byte(hash)) == 0, nil
}
