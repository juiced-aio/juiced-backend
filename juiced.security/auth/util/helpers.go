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
		return ERROR_AUTHENTICATE_DECRYPT_RESPONSE, err
	}

	if !authenticateResponse.Success {
		if authenticateResponse.ErrorMessage == "Token expired" {
			return ERROR_AUTHENTICATE_TOKEN_EXPIRED, errors.New("token expired")
		}
		return ERROR_AUTHENTICATE_FAILED, errors.New(authenticateResponse.ErrorMessage)
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
		return userInfo, ERROR_REFRESH_DECRYPT_RESPONSE, err
	}

	if !refreshResponse.Success {
		return userInfo, ERROR_REFRESH_FAILED, errors.New(refreshResponse.ErrorMessage)
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

func PX(site, proxy string, userInfo entities.UserInfo) (PXAPIResponse, PXResult, error) {
	pxAPIResponse := PXAPIResponse{}
	pxResponse := PXResponse{}
	encryptedPXResponse := EncryptedPXResponse{}

	endpoint := "https://identity.juicedbot.io/api/v1/juiced/p"
	// endpoint := "http://127.0.0.1:5000/api/v1/juiced/p"

	hwid, err := machineid.ProtectedID("juiced")
	if err != nil {
		return pxAPIResponse, ERROR_PX_HWID, err
	}

	bIV := make([]byte, aes.BlockSize)
	if _, err := rand.Read(bIV); err != nil {
		return pxAPIResponse, ERROR_PX_CREATE_IV, err
	}

	key := PX_ENCRYPTION_KEY

	timestamp := time.Now().Unix()
	encryptedTimestamp, err := Aes256Encrypt(userInfo.Email+"|JUICED|"+fmt.Sprint(timestamp), key)
	if err != nil {
		return pxAPIResponse, ERROR_PX_ENCRYPT_TIMESTAMP, err
	}

	key = strings.Replace(key, key[:len(fmt.Sprint(timestamp))], fmt.Sprint(timestamp), 1)

	encryptedActivationToken, err := Aes256Encrypt(userInfo.ActivationToken, key)
	if err != nil {
		return pxAPIResponse, ERROR_PX_ENCRYPT_ACTIVATION_TOKEN, err
	}
	encryptedHWID, err := Aes256Encrypt(hwid, key)
	if err != nil {
		return pxAPIResponse, ERROR_PX_ENCRYPT_HWID, err
	}
	encryptedDeviceName, err := Aes256Encrypt(userInfo.DeviceName, key)
	if err != nil {
		return pxAPIResponse, ERROR_PX_ENCRYPT_DEVICE_NAME, err
	}
	encryptedSite, err := Aes256Encrypt(site, key)
	if err != nil {
		return pxAPIResponse, ERROR_PX_ENCRYPT_DEVICE_NAME, err
	}
	encryptedProxy, err := Aes256Encrypt(proxy, key)
	if err != nil {
		return pxAPIResponse, ERROR_PX_ENCRYPT_DEVICE_NAME, err
	}

	encryptedHeaderA, err := Aes256Encrypt(userInfo.LicenseKey[:3], key)
	if err != nil {
		return pxAPIResponse, ERROR_PX_ENCRYPT_HEADER_A, err
	}
	encryptedHeaderE, err := Aes256Encrypt(userInfo.LicenseKey[3:4], key)
	if err != nil {
		return pxAPIResponse, ERROR_PX_ENCRYPT_HEADER_E, err
	}
	encryptedHeaderB, err := Aes256Encrypt(userInfo.LicenseKey[4:14], key)
	if err != nil {
		return pxAPIResponse, ERROR_PX_ENCRYPT_HEADER_B, err
	}
	encryptedHeaderD, err := Aes256Encrypt(userInfo.LicenseKey[14:16], key)
	if err != nil {
		return pxAPIResponse, ERROR_PX_ENCRYPT_HEADER_D, err
	}
	encryptedHeaderC, err := Aes256Encrypt(userInfo.LicenseKey[16:], key)
	if err != nil {
		return pxAPIResponse, ERROR_PX_ENCRYPT_HEADER_C, err
	}

	pxRequest := PXRequest{
		HWID:            encryptedHWID,
		DeviceName:      encryptedDeviceName,
		ActivationToken: encryptedActivationToken,
		Site:            encryptedSite,
		Proxy:           encryptedProxy,
	}

	data, _ := json.Marshal(pxRequest)
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
		return pxAPIResponse, ERROR_PX_REQUEST, err
	}

	defer response.Body.Close()
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return pxAPIResponse, ERROR_PX_READ_BODY, err
	}

	json.Unmarshal(body, &encryptedPXResponse)

	pxResponse, err = DecryptPXResponse(encryptedPXResponse, timestamp)
	if err != nil {
		return pxAPIResponse, ERROR_PX_DECRYPT_RESPONSE, err
	}

	if !pxResponse.Success {
		return pxAPIResponse, ERROR_PX_FAILED, errors.New(pxResponse.ErrorMessage)
	}

	return pxResponse.PXAPIResponse, SUCCESS_PX, nil
}

func DecryptPXResponse(response EncryptedPXResponse, timestamp int64) (PXResponse, error) {
	pxResponse := PXResponse{}

	key := PX_DECRYPTION_KEY

	success, err := Aes256Decrypt(response.Success, key)
	if err != nil {
		return pxResponse, err
	}
	errorMessage, err := Aes256Decrypt(response.ErrorMessage, key)
	if err != nil {
		return pxResponse, err
	}

	setID, err := Aes256Decrypt(response.PXAPIResponse.SetID, key)
	if err != nil {
		return pxResponse, err
	}
	uuid, err := Aes256Decrypt(response.PXAPIResponse.UUID, key)
	if err != nil {
		return pxResponse, err
	}
	vid, err := Aes256Decrypt(response.PXAPIResponse.VID, key)
	if err != nil {
		return pxResponse, err
	}
	userAgent, err := Aes256Decrypt(response.PXAPIResponse.UserAgent, key)
	if err != nil {
		return pxResponse, err
	}
	px3, err := Aes256Decrypt(response.PXAPIResponse.PX3, key)
	if err != nil {
		return pxResponse, err
	}

	pxResponse = PXResponse{
		Success: success == "true",
		PXAPIResponse: PXAPIResponse{
			SetID:     setID,
			UUID:      uuid,
			VID:       vid,
			UserAgent: userAgent,
			PX3:       px3,
		},
		ErrorMessage: errorMessage,
	}

	return pxResponse, nil
}

func PXCap(site, proxy, setID, vid, uuid, token string, userInfo entities.UserInfo) (string, PXCapResult, error) {
	pxCapResponse := PXCapResponse{}
	encryptedPXCapResponse := EncryptedPXCapResponse{}

	endpoint := "https://identity.juicedbot.io/api/v1/juiced/pc"
	// endpoint := "http://127.0.0.1:5000/api/v1/juiced/pc"

	hwid, err := machineid.ProtectedID("juiced")
	if err != nil {
		return "", ERROR_PX_CAP_HWID, err
	}

	bIV := make([]byte, aes.BlockSize)
	if _, err := rand.Read(bIV); err != nil {
		return "", ERROR_PX_CAP_CREATE_IV, err
	}

	key := PXCAP_ENCRYPTION_KEY

	timestamp := time.Now().Unix()
	encryptedTimestamp, err := Aes256Encrypt(userInfo.Email+"|JUICED|"+fmt.Sprint(timestamp), key)
	if err != nil {
		return "", ERROR_PX_CAP_ENCRYPT_TIMESTAMP, err
	}

	key = strings.Replace(key, key[:len(fmt.Sprint(timestamp))], fmt.Sprint(timestamp), 1)

	encryptedActivationToken, err := Aes256Encrypt(userInfo.ActivationToken, key)
	if err != nil {
		return "", ERROR_PX_CAP_ENCRYPT_ACTIVATION_TOKEN, err
	}
	encryptedHWID, err := Aes256Encrypt(hwid, key)
	if err != nil {
		return "", ERROR_PX_CAP_ENCRYPT_HWID, err
	}
	encryptedDeviceName, err := Aes256Encrypt(userInfo.DeviceName, key)
	if err != nil {
		return "", ERROR_PX_CAP_ENCRYPT_DEVICE_NAME, err
	}
	encryptedSite, err := Aes256Encrypt(site, key)
	if err != nil {
		return "", ERROR_PX_CAP_ENCRYPT_DEVICE_NAME, err
	}
	encryptedProxy, err := Aes256Encrypt(proxy, key)
	if err != nil {
		return "", ERROR_PX_CAP_ENCRYPT_DEVICE_NAME, err
	}
	encryptedSetID, err := Aes256Encrypt(setID, key)
	if err != nil {
		return "", ERROR_PX_CAP_ENCRYPT_DEVICE_NAME, err
	}
	encryptedUUID, err := Aes256Encrypt(uuid, key)
	if err != nil {
		return "", ERROR_PX_CAP_ENCRYPT_DEVICE_NAME, err
	}
	encryptedVID, err := Aes256Encrypt(vid, key)
	if err != nil {
		return "", ERROR_PX_CAP_ENCRYPT_DEVICE_NAME, err
	}
	encryptedToken, err := Aes256Encrypt(token, key)
	if err != nil {
		return "", ERROR_PX_CAP_ENCRYPT_DEVICE_NAME, err
	}

	encryptedHeaderE, err := Aes256Encrypt(userInfo.LicenseKey[:3], key)
	if err != nil {
		return "", ERROR_PX_CAP_ENCRYPT_HEADER_E, err
	}
	encryptedHeaderA, err := Aes256Encrypt(userInfo.LicenseKey[3:4], key)
	if err != nil {
		return "", ERROR_PX_CAP_ENCRYPT_HEADER_A, err
	}
	encryptedHeaderD, err := Aes256Encrypt(userInfo.LicenseKey[4:14], key)
	if err != nil {
		return "", ERROR_PX_CAP_ENCRYPT_HEADER_D, err
	}
	encryptedHeaderB, err := Aes256Encrypt(userInfo.LicenseKey[14:19], key)
	if err != nil {
		return "", ERROR_PX_CAP_ENCRYPT_HEADER_B, err
	}
	encryptedHeaderC, err := Aes256Encrypt(userInfo.LicenseKey[19:], key)
	if err != nil {
		return "", ERROR_PX_CAP_ENCRYPT_HEADER_C, err
	}

	pxCapRequest := PXCapRequest{
		HWID:            encryptedHWID,
		DeviceName:      encryptedDeviceName,
		ActivationToken: encryptedActivationToken,
		Site:            encryptedSite,
		Proxy:           encryptedProxy,
		SetID:           encryptedSetID,
		UUID:            encryptedUUID,
		VID:             encryptedVID,
		Token:           encryptedToken,
	}

	data, _ := json.Marshal(pxCapRequest)
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
		return "", ERROR_PX_CAP_REQUEST, err
	}

	defer response.Body.Close()
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return "", ERROR_PX_CAP_READ_BODY, err
	}

	json.Unmarshal(body, &pxCapResponse)

	pxCapResponse, err = DecryptPXCapResponse(encryptedPXCapResponse, timestamp)
	if err != nil {
		return "", ERROR_PX_CAP_DECRYPT_RESPONSE, err
	}

	if !pxCapResponse.Success {
		return "", ERROR_PX_CAP_FAILED, errors.New(pxCapResponse.ErrorMessage)
	}

	return pxCapResponse.PX3, SUCCESS_PX_CAP, nil
}

func DecryptPXCapResponse(response EncryptedPXCapResponse, timestamp int64) (PXCapResponse, error) {
	pxCapResponse := PXCapResponse{}

	key := PXCAP_DECRYPTION_KEY

	success, err := Aes256Decrypt(response.Success, key)
	if err != nil {
		return pxCapResponse, err
	}
	errorMessage, err := Aes256Decrypt(response.ErrorMessage, key)
	if err != nil {
		return pxCapResponse, err
	}
	px3, err := Aes256Decrypt(response.PX3, key)
	if err != nil {
		return pxCapResponse, err
	}

	pxCapResponse = PXCapResponse{
		Success:      success == "true",
		ErrorMessage: errorMessage,
		PX3:          px3,
	}

	return pxCapResponse, nil
}

func Akamai(pageURL, skipKact, skipMact, onBlur, onFocus, abck, sensorDataLink, ver, firstPost, pixelID, pixelG, json_ string, userInfo entities.UserInfo) (AkamaiAPIResponse, AkamaiResult, error) {
	akamaiAPIResponse := AkamaiAPIResponse{}
	akamaiResponse := AkamaiResponse{}
	encryptedAkamaiResponse := EncryptedAkamaiResponse{}

	endpoint := "https://identity.juicedbot.io/api/v1/juiced/ak"
	// endpoint := "http://127.0.0.1:5000/api/v1/juiced/ak"

	hwid, err := machineid.ProtectedID("juiced")
	if err != nil {
		return akamaiAPIResponse, ERROR_AKAMAI_HWID, err
	}

	bIV := make([]byte, aes.BlockSize)
	if _, err := rand.Read(bIV); err != nil {
		return akamaiAPIResponse, ERROR_AKAMAI_CREATE_IV, err
	}

	key := AKAMAI_ENCRYPTION_KEY

	timestamp := time.Now().Unix()
	encryptedTimestamp, err := Aes256Encrypt(userInfo.Email+"|JUICED|"+fmt.Sprint(timestamp), key)
	if err != nil {
		return akamaiAPIResponse, ERROR_AKAMAI_ENCRYPT_TIMESTAMP, err
	}

	key = strings.Replace(key, key[:len(fmt.Sprint(timestamp))], fmt.Sprint(timestamp), 1)

	encryptedActivationToken, err := Aes256Encrypt(userInfo.ActivationToken, key)
	if err != nil {
		return akamaiAPIResponse, ERROR_AKAMAI_ENCRYPT_ACTIVATION_TOKEN, err
	}
	encryptedHWID, err := Aes256Encrypt(hwid, key)
	if err != nil {
		return akamaiAPIResponse, ERROR_AKAMAI_ENCRYPT_HWID, err
	}
	encryptedDeviceName, err := Aes256Encrypt(userInfo.DeviceName, key)
	if err != nil {
		return akamaiAPIResponse, ERROR_AKAMAI_ENCRYPT_DEVICE_NAME, err
	}

	encryptedPageURL, err := Aes256Encrypt(pageURL, key)
	if err != nil {
		return akamaiAPIResponse, ERROR_AKAMAI_ENCRYPT_PAGE_URL, err
	}
	encryptedSkipKact, err := Aes256Encrypt(skipKact, key)
	if err != nil {
		return akamaiAPIResponse, ERROR_AKAMAI_ENCRYPT_SKIP_KACT, err
	}
	encryptedSkipMact, err := Aes256Encrypt(skipMact, key)
	if err != nil {
		return akamaiAPIResponse, ERROR_AKAMAI_ENCRYPT_SKIP_MACT, err
	}
	encryptedOnBlur, err := Aes256Encrypt(onBlur, key)
	if err != nil {
		return akamaiAPIResponse, ERROR_AKAMAI_ENCRYPT_ON_BLUR, err
	}
	encryptedOnFocus, err := Aes256Encrypt(onFocus, key)
	if err != nil {
		return akamaiAPIResponse, ERROR_AKAMAI_ENCRYPT_ON_FOCUS, err
	}
	encryptedAbck, err := Aes256Encrypt(abck, key)
	if err != nil {
		return akamaiAPIResponse, ERROR_AKAMAI_ENCRYPT_ABCK, err
	}
	encryptedSensorDataLink, err := Aes256Encrypt(sensorDataLink, key)
	if err != nil {
		return akamaiAPIResponse, ERROR_AKAMAI_ENCRYPT_SENSOR_DATA_LINK, err
	}
	encryptedVer, err := Aes256Encrypt(ver, key)
	if err != nil {
		return akamaiAPIResponse, ERROR_AKAMAI_ENCRYPT_VER, err
	}
	encryptedFirstPost, err := Aes256Encrypt(firstPost, key)
	if err != nil {
		return akamaiAPIResponse, ERROR_AKAMAI_ENCRYPT_FIRST_POST, err
	}
	encryptedPixelID, err := Aes256Encrypt(pixelID, key)
	if err != nil {
		return akamaiAPIResponse, ERROR_AKAMAI_ENCRYPT_PIXEL_ID, err
	}
	encryptedPixelG, err := Aes256Encrypt(pixelG, key)
	if err != nil {
		return akamaiAPIResponse, ERROR_AKAMAI_ENCRYPT_PIXEL_G, err
	}
	encryptedJSON, err := Aes256Encrypt(json_, key)
	if err != nil {
		return akamaiAPIResponse, ERROR_AKAMAI_ENCRYPT_JSON, err
	}

	encryptedHeaderA, err := Aes256Encrypt(userInfo.LicenseKey[:3], key)
	if err != nil {
		return akamaiAPIResponse, ERROR_AKAMAI_ENCRYPT_HEADER_B, err
	}
	encryptedHeaderD, err := Aes256Encrypt(userInfo.LicenseKey[3:7], key)
	if err != nil {
		return akamaiAPIResponse, ERROR_AKAMAI_ENCRYPT_HEADER_C, err
	}
	encryptedHeaderE, err := Aes256Encrypt(userInfo.LicenseKey[7:14], key)
	if err != nil {
		return akamaiAPIResponse, ERROR_AKAMAI_ENCRYPT_HEADER_A, err
	}
	encryptedHeaderB, err := Aes256Encrypt(userInfo.LicenseKey[14:18], key)
	if err != nil {
		return akamaiAPIResponse, ERROR_AKAMAI_ENCRYPT_HEADER_E, err
	}
	encryptedHeaderC, err := Aes256Encrypt(userInfo.LicenseKey[18:], key)
	if err != nil {
		return akamaiAPIResponse, ERROR_AKAMAI_ENCRYPT_HEADER_D, err
	}

	akamaiRequest := AkamaiRequest{
		HWID:            encryptedHWID,
		DeviceName:      encryptedDeviceName,
		ActivationToken: encryptedActivationToken,
		PageURL:         encryptedPageURL,
		SkipKact:        encryptedSkipKact,
		SkipMact:        encryptedSkipMact,
		OnBlur:          encryptedOnBlur,
		OnFocus:         encryptedOnFocus,
		Abck:            encryptedAbck,
		SensorDataLink:  encryptedSensorDataLink,
		Ver:             encryptedVer,
		FirstPost:       encryptedFirstPost,
		PixelID:         encryptedPixelID,
		PixelG:          encryptedPixelG,
		JSON:            encryptedJSON,
	}

	data, _ := json.Marshal(akamaiRequest)
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
		return akamaiAPIResponse, ERROR_AKAMAI_REQUEST, err
	}

	defer response.Body.Close()
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return akamaiAPIResponse, ERROR_AKAMAI_READ_BODY, err
	}

	json.Unmarshal(body, &encryptedAkamaiResponse)

	akamaiResponse, err = DecryptAkamaiResponse(encryptedAkamaiResponse, timestamp)
	if err != nil {
		return akamaiAPIResponse, ERROR_AKAMAI_DECRYPT_RESPONSE, err
	}

	if !akamaiResponse.Success {
		return akamaiAPIResponse, ERROR_AKAMAI_FAILED, errors.New(akamaiResponse.ErrorMessage)
	}

	return akamaiResponse.AkamaiAPIResponse, SUCCESS_AKAMAI, nil
}

func DecryptAkamaiResponse(response EncryptedAkamaiResponse, timestamp int64) (AkamaiResponse, error) {
	akamaiResponse := AkamaiResponse{}

	key := AKAMAI_DECRYPTION_KEY

	success, err := Aes256Decrypt(response.Success, key)
	if err != nil {
		return akamaiResponse, err
	}
	errorMessage, err := Aes256Decrypt(response.ErrorMessage, key)
	if err != nil {
		return akamaiResponse, err
	}
	sensorData, err := Aes256Decrypt(response.AkamaiAPIResponse.SensorData, key)
	if err != nil {
		return akamaiResponse, err
	}
	pixel, err := Aes256Decrypt(response.AkamaiAPIResponse.Pixel, key)
	if err != nil {
		return akamaiResponse, err
	}

	akamaiResponse = AkamaiResponse{
		Success:      success == "true",
		ErrorMessage: errorMessage,
		AkamaiAPIResponse: AkamaiAPIResponse{
			SensorData: sensorData,
			Pixel:      pixel,
		},
	}

	return akamaiResponse, nil
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
