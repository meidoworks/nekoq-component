package random

import (
	"math/rand"
	"time"
)

func String(length int) string {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	// String
	charset := "abcdefghijklmnopqrstuvwxyz"

	// Getting random character
	c := charset[r.Intn(len(charset))]

	// Return the character
	return string(c)
}
