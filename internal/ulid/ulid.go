package ulid

import (
	"crypto/rand"
	"time"

	"github.com/oklog/ulid"
)

func GenerateULID() (string, error) {
    entropy := rand.Reader
    ms := ulid.Timestamp(time.Now())
	id, err := ulid.New(ms, entropy)


	if err != nil {
		return "", err
	}
	return id.String(), nil
}
