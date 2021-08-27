package helpers

import (
	"errors"
	"fmt"
	"math/rand"
	"regexp"
	"time"
)

var seededRand *rand.Rand = rand.New(rand.NewSource(time.Now().UnixNano()))
var runes = []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
var runesWithLower = []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz123456789")

// RandID returns a random n-digit ID of digits and uppercase letters
func RandID(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = runes[seededRand.Intn(len(runes))]
	}
	return string(b)
}

// RandString returns a random n-digit string of digits and uppercase/lowercase letters
func RandString(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = runesWithLower[seededRand.Intn(len(runesWithLower))]
	}
	return string(b)
}

// Returns true if it finds the string x in the slice s
func InSlice(s []string, x string) bool {
	for _, i := range s {
		if i == x {
			return true
		}
	}
	return false
}

// Removes the string x from the slice s
func RemoveFromSlice(s []string, x string) []string {
	if !InSlice(s, x) {
		return s
	}
	var position int
	for i, r := range s {
		if r == x {
			position = i
			break
		}

	}
	return append(s[:position], s[position+1:]...)

}

func FindInString(str string, start string, end string) (string, error) {
	comp := regexp.MustCompile(fmt.Sprintf("%v(.*?)%v", start, end))
	comp.MatchString(str)

	o := comp.FindStringSubmatch(str)
	if len(o) == 0 {
		return "", errors.New("string not found")
	}
	parsed := o[1]
	if parsed == "" {
		return parsed, errors.New("string not found")
	}

	return parsed, nil
}
