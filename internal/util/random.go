package util

import (
	"math/rand"
	"strings"
	"time"
)

const alphabet = "abcdefghijklmnopqrstuvwxyz"

var seededRand = rand.New(rand.NewSource(time.Now().UnixNano()))

// RandomInt generates a random integer between min and max
func RandomInt(min, max int32) int32 {
	return min + seededRand.Int31n(max-min+1)
}

// RandomString generates a random sting of length n
func RandomString(length int) string {
	var sb strings.Builder
	k := len(alphabet)
	for i := 0; i < length; i++ {
		c := alphabet[seededRand.Intn(k)]
		sb.WriteByte(c)
	}
	return sb.String()
}

// RandomName generates  random names
func RandomName() string {
	return RandomString(10)
}

// RandomEmail generates a random email address
func RandomEmail() string {
	return RandomString(10) + "@gmail.com"
}

// RandomPhone generates a random 10-digit phone number
func RandomPhone() string {
	var sb strings.Builder
	for i := 0; i < 10; i++ {
		digit := RandomInt(0, 9)
		sb.WriteByte(byte('0' + digit)) // convert int to ASCII character
	}
	return sb.String()
}

func RandomAge() int32 {
	return RandomInt(1, 120)
}

// RandomStatus randomly returns "Active" or "Inactive"
func RandomStatus() string {
	statuses := []string{"Active", "Inactive"}
	n := seededRand.Intn(len(statuses)) // random 0 or 1
	return statuses[n]
}
