package misc

import (
	"math/rand"
	"time"
)

func RandomStringBytes(length int) string {
	letter := []rune("abcdefghijklmnopqrstuvwxyz0123456789")
	b := make([]rune, length)
	rand.Seed(time.Now().UnixNano())

	for i := range b {
		b[i] = letter[rand.Intn(len(letter))]
	}
	return string(b)
}

