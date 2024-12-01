package random

import (
	"math/rand"
	"time"
)

func String(length int) string {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	ran_str := make([]byte, length)

	// String
	charset := "abcdefghijklmnopqrstuvwxyz"

	// Getting random character
	for i := 0; i < length; i++ {
		ran_str[i] = charset[r.Intn(len(charset))]
	}

	// Return the character
	return string(ran_str)
}
