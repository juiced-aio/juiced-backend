package util

import (
	"crypto/aes"
	"encoding/hex"
	"reflect"
	"testing"
)

func TestAes256Encrypt(t *testing.T) {
	tables := []struct {
		plaintext string
		key       string
		errorType reflect.Type
	}{
		// Case 1 (pkcs7.Pad err) -- even though pkcs7.Pad returns an error value, it is always nil
		{"StringToEncrypt", "INVALID_KEY_LENGTH", reflect.TypeOf(aes.KeySizeError(0))}, // Case 2 (aes.NewCipher err): Key length is not equal to 16, 24, or 32
		// Case 3 (rand.Read err) -- cannot force a rand.Read error
		{"StringToEncrypt", "VALID_KEY_LENGTH", nil}, // Case 4: Success
	}

	for _, table := range tables {
		_, err := Aes256Encrypt(table.plaintext, table.key)
		if reflect.TypeOf(err) != table.errorType {
			t.Errorf("Aes256Encrypt of `%s` with key `%s` was incorrect, got: err = `%s`, want: err = `%s`.", table.plaintext, table.key, reflect.TypeOf(err).String(), table.errorType.String())
		}
	}
}

func TestAes256Decrypt(t *testing.T) {
	tables := []struct {
		encryptFirst  bool
		plaintext     string
		encryptedText string
		key           string
		errorType     reflect.Type
	}{
		{false, "", "OddEncryptedValue", "", reflect.TypeOf(hex.InvalidByteError([]byte("0")[0]))},                                                 // Case 1 (hex.DecodeString err): Length of encrypted text is odd (Aes256Encrypt always creates an even-length encryption string)
		{false, "", "b575b1de7057d46c156a0ada54384e1f329cdcddf20f74b5356836cef7892883", "INVALID_KEY_LENGTH", reflect.TypeOf(aes.KeySizeError(0))}, // Case 2 (aes.NewCipher err): Key length is not equal to 16, 24, or 32
		// TODO: Case 3 (CipherTextTooShortError)
		// TODO: Case 4 (CipherTextNotMultipleOfBlockSizeError)
		// TODO: Case 5 (pkcs7.Unpad err)
		// TODO: Case 6 (Success)
	}

	for _, table := range tables {
		encryptedText := table.encryptedText
		var err error
		if table.encryptFirst {
			encryptedText, err = Aes256Encrypt(table.plaintext, table.key)
		}
		if err != nil {
			// We've already tested Aes256Encrypt, so we should be passing values here that we expect to succeed.
			// If err != nil at this point, then we have an issue.
			t.Errorf("Aes256Decrypt failed due to Aes256Encrypt failing: err %s", err.Error())
		}
		decryptedText, err := Aes256Decrypt(encryptedText, table.key)
		if reflect.TypeOf(err) != table.errorType || decryptedText != table.plaintext {
			t.Errorf("Aes256Decrypt of `%s` with key `%s` was incorrect, got: plaintext = `%s` / err = `%s`, want: plaintext = `%s` / err = `%s`.", table.plaintext, table.key, decryptedText, reflect.TypeOf(err).String(), table.plaintext, table.errorType.String())
		}
	}
}
