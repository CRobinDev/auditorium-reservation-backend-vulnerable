package randgen

import (
	"math"
	"math/rand"
)

func RandomNumber(digits int) int {
	low := int(math.Pow10(digits - 1))
	high := int(math.Pow10(digits)) - 1
	return low + rand.Intn(high-low+1)
}

func RandomString(length int) string {
	alphaNumRunes := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890")
	randomRune := make([]rune, length)

	for i := 0; i < length; i++ {
		randomRune[i] = alphaNumRunes[rand.Intn(len(alphaNumRunes)-1)]
	}

	return string(randomRune)
}
