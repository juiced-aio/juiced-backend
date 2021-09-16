package util

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/hex"
	"errors"
	"fmt"
	"math/rand"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/mergermarket/go-pkcs7"
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

func FindInString2(str string, start string, end string) (string, error) {
	if !strings.Contains(str, start) {
		return "", errors.New("string not found")
	}

	after := strings.Split(str, start)
	if len(after) < 2 || after[1] == "" {
		return "", errors.New("string not found")
	}

	if !strings.Contains(after[1], end) {
		return "", errors.New("string not found")
	}

	between := strings.Split(after[1], end)
	if len(between) == 0 || between[0] == "" {
		return "", errors.New("string not found")
	}

	return between[0], nil
}

func CreateParams(paramsLong map[string]string) string {
	params := url.Values{}
	for key, value := range paramsLong {
		params.Add(key, value)
	}
	return params.Encode()
}

func Aes256Encrypt(plaintext string, key string) (string, error) {
	bKey := []byte(key)
	bPlaintext, err := pkcs7.Pad([]byte(plaintext), aes.BlockSize)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(bKey)
	if err != nil {
		return "", err
	}
	cipherText := make([]byte, aes.BlockSize+len(bPlaintext))
	bIV := cipherText[:aes.BlockSize]
	if _, err := rand.Read(bIV); err != nil {
		return "", err
	}

	mode := cipher.NewCBCEncrypter(block, bIV)
	mode.CryptBlocks(cipherText[aes.BlockSize:], bPlaintext)
	return fmt.Sprintf("%x", cipherText), nil
}

func Aes256Decrypt(encryptedText string, key string) (string, error) {
	bKey := []byte(key)
	cipherText, err := hex.DecodeString(encryptedText)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(bKey)
	if err != nil {
		return "", err
	}

	if len(cipherText) < aes.BlockSize {
		return "", &CipherTextTooShortError{}
	}
	iv := cipherText[:aes.BlockSize]
	cipherText = cipherText[aes.BlockSize:]
	if len(cipherText)%aes.BlockSize != 0 {
		return "", &CipherTextNotMultipleOfBlockSizeError{}
	}

	mode := cipher.NewCBCDecrypter(block, iv)
	mode.CryptBlocks(cipherText, cipherText)

	cipherText, err = pkcs7.Unpad(cipherText, aes.BlockSize)
	return string(cipherText), err
}

func EncryptValues(key string, values ...string) (encryptedValues []string, _ error) {
	for _, value := range values {
		e, err := Aes256Encrypt(value, key)
		if err != nil {
			return encryptedValues, err
		}
		encryptedValues = append(encryptedValues, e)
	}
	return encryptedValues, nil
}
