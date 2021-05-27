package util

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"

	"backend.juicedbot.io/m/v2/juiced.infrastructure/commands"
	"backend.juicedbot.io/m/v2/juiced.infrastructure/common/entities"

	"github.com/denisbrodbeck/machineid"
	"github.com/mergermarket/go-pkcs7"
)

func Authenticate(userInfo entities.UserInfo) (AuthenticationResult, error) {
	authenticateResponse := AuthenticateResponse{}
	encryptedAuthenticateResponse := EncryptedAuthenticateResponse{}

	endpoint := "https://identity.juicedbot.io/api/v1/juiced/a"
	// endpoint := "http://127.0.0.1:5000/api/v1/juiced/a"

	hwid, err := machineid.ProtectedID("juiced")
	if err != nil {
		return ERROR_AUTHENTICATE_HWID, err
	}

	bIV := make([]byte, aes.BlockSize)
	if _, err := rand.Read(bIV); err != nil {
		return ERROR_AUTHENTICATE_CREATE_IV, err
	}

	key := AUTHENTICATION_ENCRYPTION_KEY

	timestamp := time.Now().Unix()
	encryptedTimestamp, err := Aes256Encrypt(userInfo.Email+"|JUICED|"+fmt.Sprint(timestamp), key)
	if err != nil {
		return ERROR_AUTHENTICATE_ENCRYPT_TIMESTAMP, err
	}

	key = strings.Replace(key, key[:len(fmt.Sprint(timestamp))], fmt.Sprint(timestamp), 1)

	encryptedActivationToken, err := Aes256Encrypt(userInfo.ActivationToken, key)
	if err != nil {
		return ERROR_AUTHENTICATE_ENCRYPT_ACTIVATION_TOKEN, err
	}
	encryptedHWID, err := Aes256Encrypt(hwid, key)
	if err != nil {
		return ERROR_AUTHENTICATE_ENCRYPT_HWID, err
	}
	encryptedDeviceName, err := Aes256Encrypt(userInfo.DeviceName, key)
	if err != nil {
		return ERROR_AUTHENTICATE_ENCRYPT_DEVICE_NAME, err
	}

	encryptedHeaderB, err := Aes256Encrypt(userInfo.LicenseKey[:4], key)
	if err != nil {
		return ERROR_AUTHENTICATE_ENCRYPT_HEADER_B, err
	}
	encryptedHeaderC, err := Aes256Encrypt(userInfo.LicenseKey[4:10], key)
	if err != nil {
		return ERROR_AUTHENTICATE_ENCRYPT_HEADER_C, err
	}
	encryptedHeaderA, err := Aes256Encrypt(userInfo.LicenseKey[10:15], key)
	if err != nil {
		return ERROR_AUTHENTICATE_ENCRYPT_HEADER_A, err
	}
	encryptedHeaderE, err := Aes256Encrypt(userInfo.LicenseKey[15:19], key)
	if err != nil {
		return ERROR_AUTHENTICATE_ENCRYPT_HEADER_E, err
	}
	encryptedHeaderD, err := Aes256Encrypt(userInfo.LicenseKey[19:], key)
	if err != nil {
		return ERROR_AUTHENTICATE_ENCRYPT_HEADER_D, err
	}

	authenticateRequest := AuthenticateRequest{
		HWID:            encryptedHWID,
		DeviceName:      encryptedDeviceName,
		ActivationToken: encryptedActivationToken,
	}

	data, _ := json.Marshal(authenticateRequest)
	payload := bytes.NewBuffer(data)
	request, _ := http.NewRequest("POST", endpoint, payload)
	request.Header.Add("x-j-w", encryptedTimestamp)
	request.Header.Add("x-j-a", encryptedHeaderA)
	request.Header.Add("x-j-b", encryptedHeaderB)
	request.Header.Add("x-j-c", encryptedHeaderC)
	request.Header.Add("x-j-d", encryptedHeaderD)
	request.Header.Add("x-j-e", encryptedHeaderE)
	request.Header.Add("Content-Type", "application/json")
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return ERROR_AUTHENTICATE_REQUEST, err
	}

	defer response.Body.Close()
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return ERROR_AUTHENTICATE_READ_BODY, err
	}

	json.Unmarshal(body, &encryptedAuthenticateResponse)

	authenticateResponse, err = DecryptAuthenticateResponse(encryptedAuthenticateResponse, timestamp)
	if err != nil {
		return ERROR_AUTHENTICATE_DECRYPT_RESPONSE, errors.New("")
	}

	if !authenticateResponse.Success {
		if authenticateResponse.ErrorMessage == "Token expired" {
			return ERROR_AUTHENTICATE_TOKEN_EXPIRED, errors.New("token expired")
		}
		return ERROR_AUTHENTICATE_FAILED, errors.New("invalid user info")
	}

	return SUCCESS_AUTHENTICATE, nil
}

func DecryptAuthenticateResponse(response EncryptedAuthenticateResponse, timestamp int64) (AuthenticateResponse, error) {
	authenticateResponse := AuthenticateResponse{}

	key := AUTHENTICATION_DECRYPTION_KEY

	success, err := Aes256Decrypt(response.Success, key)
	if err != nil {
		return authenticateResponse, err
	}

	errorMessage, err := Aes256Decrypt(response.ErrorMessage, key)
	if err != nil {
		return authenticateResponse, err
	}

	authenticateResponse = AuthenticateResponse{
		Success:      success == "true",
		ErrorMessage: errorMessage,
	}

	return authenticateResponse, nil
}

func Refresh(userInfo entities.UserInfo) (entities.UserInfo, RefreshResult, error) {
	refreshResponse := RefreshTokenResponse{}
	encryptedRefreshResponse := EncryptedRefreshTokenResponse{}

	endpoint := "https://identity.juicedbot.io/api/v1/juiced/r"
	// endpoint := "http://127.0.0.1:5000/api/v1/juiced/r"

	hwid, err := machineid.ProtectedID("juiced")
	if err != nil {
		return userInfo, ERROR_REFRESH_HWID, err
	}

	bIV := make([]byte, aes.BlockSize)
	if _, err := rand.Read(bIV); err != nil {
		return userInfo, ERROR_REFRESH_CREATE_IV, err
	}

	key := REFRESH_ENCRYPTION_KEY

	timestamp := time.Now().Unix()
	encryptedTimestamp, err := Aes256Encrypt(userInfo.Email+"|JUICED|"+fmt.Sprint(timestamp), key)
	if err != nil {
		return userInfo, ERROR_REFRESH_ENCRYPT_TIMESTAMP, err
	}

	key = strings.Replace(key, key[:len(fmt.Sprint(timestamp))], fmt.Sprint(timestamp), 1)

	encryptedActivationToken, err := Aes256Encrypt(userInfo.ActivationToken, key)
	if err != nil {
		return userInfo, ERROR_REFRESH_ENCRYPT_ACTIVATION_TOKEN, err
	}
	encryptedRefreshToken, err := Aes256Encrypt(userInfo.RefreshToken, key)
	if err != nil {
		return userInfo, ERROR_REFRESH_ENCRYPT_REFRESH_TOKEN, err
	}
	encryptedHWID, err := Aes256Encrypt(hwid, key)
	if err != nil {
		return userInfo, ERROR_REFRESH_ENCRYPT_HWID, err
	}
	encryptedDeviceName, err := Aes256Encrypt(userInfo.DeviceName, key)
	if err != nil {
		return userInfo, ERROR_REFRESH_ENCRYPT_DEVICE_NAME, err
	}

	encryptedHeaderA, err := Aes256Encrypt(userInfo.LicenseKey[:3], key)
	if err != nil {
		return userInfo, ERROR_REFRESH_ENCRYPT_HEADER_B, err
	}
	encryptedHeaderD, err := Aes256Encrypt(userInfo.LicenseKey[3:7], key)
	if err != nil {
		return userInfo, ERROR_REFRESH_ENCRYPT_HEADER_C, err
	}
	encryptedHeaderE, err := Aes256Encrypt(userInfo.LicenseKey[7:14], key)
	if err != nil {
		return userInfo, ERROR_REFRESH_ENCRYPT_HEADER_A, err
	}
	encryptedHeaderB, err := Aes256Encrypt(userInfo.LicenseKey[14:18], key)
	if err != nil {
		return userInfo, ERROR_REFRESH_ENCRYPT_HEADER_E, err
	}
	encryptedHeaderC, err := Aes256Encrypt(userInfo.LicenseKey[18:], key)
	if err != nil {
		return userInfo, ERROR_REFRESH_ENCRYPT_HEADER_D, err
	}

	refreshRequest := RefreshTokenRequest{
		HWID:            encryptedHWID,
		DeviceName:      encryptedDeviceName,
		ActivationToken: encryptedActivationToken,
		RefreshToken:    encryptedRefreshToken,
	}

	data, _ := json.Marshal(refreshRequest)
	payload := bytes.NewBuffer(data)
	request, _ := http.NewRequest("POST", endpoint, payload)
	request.Header.Add("x-j-w", encryptedTimestamp)
	request.Header.Add("x-j-a", encryptedHeaderA)
	request.Header.Add("x-j-b", encryptedHeaderB)
	request.Header.Add("x-j-c", encryptedHeaderC)
	request.Header.Add("x-j-d", encryptedHeaderD)
	request.Header.Add("x-j-e", encryptedHeaderE)
	request.Header.Add("Content-Type", "application/json")
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return userInfo, ERROR_REFRESH_REQUEST, err
	}

	defer response.Body.Close()
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return userInfo, ERROR_REFRESH_READ_BODY, err
	}

	json.Unmarshal(body, &encryptedRefreshResponse)

	refreshResponse, err = DecryptRefreshResponse(encryptedRefreshResponse, timestamp)
	if err != nil {
		return userInfo, ERROR_REFRESH_DECRYPT_RESPONSE, errors.New("")
	}

	if !refreshResponse.Success {
		return userInfo, ERROR_REFRESH_FAILED, errors.New("invalid user info")
	}

	userInfo.ActivationToken = refreshResponse.ActivationToken
	userInfo.RefreshToken = refreshResponse.RefreshToken
	userInfo.ExpiresAt = refreshResponse.ExpiresAt

	err = commands.SetUserInfo(userInfo)
	if err != nil {
		return userInfo, SUCCESS_REFRESH_ERROR_SET_USER_INFO, err
	}

	return userInfo, SUCCESS_REFRESH, nil
}

func DecryptRefreshResponse(response EncryptedRefreshTokenResponse, timestamp int64) (RefreshTokenResponse, error) {
	refreshResponse := RefreshTokenResponse{}

	key := REFRESH_DECRYPTION_KEY

	success, err := Aes256Decrypt(response.Success, key)
	if err != nil {
		return refreshResponse, err
	}
	activationToken, err := Aes256Decrypt(response.ActivationToken, key)
	if err != nil {
		return refreshResponse, err
	}
	refreshToken, err := Aes256Decrypt(response.RefreshToken, key)
	if err != nil {
		return refreshResponse, err
	}
	activationTokenExpiresAt, err := Aes256Decrypt(response.ExpiresAt, key)
	if err != nil {
		return refreshResponse, err
	}
	expiresAt, err := strconv.ParseInt(activationTokenExpiresAt, 10, 64)
	if err != nil {
		return refreshResponse, err
	}
	errorMessage, err := Aes256Decrypt(response.ErrorMessage, key)
	if err != nil {
		return refreshResponse, err
	}

	refreshResponse = RefreshTokenResponse{
		Success:         success == "true",
		ErrorMessage:    errorMessage,
		ActivationToken: activationToken,
		RefreshToken:    refreshToken,
		ExpiresAt:       expiresAt,
	}

	return refreshResponse, nil
}

func Heartbeat(userInfo entities.UserInfo, retries int) (entities.UserInfo, error) {
	errCode, err := Authenticate(userInfo)
	if err == nil {
		return userInfo, nil
	}

	if errCode == ERROR_AUTHENTICATE_TOKEN_EXPIRED {
		userInfo, _, err = Refresh(userInfo)
		if err == nil {
			return userInfo, nil
		}
	}

	if retries >= MAX_RETRIES {
		return userInfo, errors.New("max retries reached")
	}
	return Heartbeat(userInfo, retries+1)
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
		return "", errors.New("cipher text is too short")
	}
	iv := cipherText[:aes.BlockSize]
	cipherText = cipherText[aes.BlockSize:]
	if len(cipherText)%aes.BlockSize != 0 {
		return "", errors.New("cipher text not a multiple of the block size")
	}

	mode := cipher.NewCBCDecrypter(block, iv)
	mode.CryptBlocks(cipherText, cipherText)

	cipherText, _ = pkcs7.Unpad(cipherText, aes.BlockSize)
	return fmt.Sprintf("%s", cipherText), nil
}
