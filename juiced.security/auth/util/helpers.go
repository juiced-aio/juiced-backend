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

	"backend.juicedbot.io/juiced.infrastructure/commands"
	"backend.juicedbot.io/juiced.infrastructure/common/entities"

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

func DiscordWebhook(success bool, content string, embeds []DiscordEmbed, userInfo entities.UserInfo) (DiscordWebhookResult, error) {
	discordWebhookResponse := DiscordWebhookResponse{}
	encryptedDiscordWebhookResponse := EncryptedDiscordWebhookResponse{}

	endpoint := "https://identity.juicedbot.io/api/v1/juiced/dw"
	// endpoint := "http://127.0.0.1:5000/api/v1/juiced/dw"

	hwid, err := machineid.ProtectedID("juiced")
	if err != nil {
		return ERROR_DISCORD_WEBHOOK_HWID, err
	}

	bIV := make([]byte, aes.BlockSize)
	if _, err := rand.Read(bIV); err != nil {
		return ERROR_DISCORD_WEBHOOK_CREATE_IV, err
	}

	key := DISCORD_WEBHOOK_ENCRYPTION_KEY

	timestamp := time.Now().Unix()
	encryptedTimestamp, err := Aes256Encrypt(userInfo.Email+"|JUICED|"+fmt.Sprint(timestamp), key)
	if err != nil {
		return ERROR_DISCORD_WEBHOOK_ENCRYPT_TIMESTAMP, err
	}

	key = strings.Replace(key, key[:len(fmt.Sprint(timestamp))], fmt.Sprint(timestamp), 1)

	encryptedActivationToken, err := Aes256Encrypt(userInfo.ActivationToken, key)
	if err != nil {
		return ERROR_DISCORD_WEBHOOK_ENCRYPT_ACTIVATION_TOKEN, err
	}
	encryptedHWID, err := Aes256Encrypt(hwid, key)
	if err != nil {
		return ERROR_DISCORD_WEBHOOK_ENCRYPT_HWID, err
	}
	encryptedDeviceName, err := Aes256Encrypt(userInfo.DeviceName, key)
	if err != nil {
		return ERROR_DISCORD_WEBHOOK_ENCRYPT_DEVICE_NAME, err
	}

	encryptedSuccess, err := Aes256Encrypt(strconv.FormatBool(success), key)
	if err != nil {
		return ERROR_DISCORD_WEBHOOK_ENCRYPT_SUCCESS, err
	}
	encryptedContent, err := Aes256Encrypt(content, key)
	if err != nil {
		return ERROR_DISCORD_WEBHOOK_ENCRYPT_CONTENT, err
	}

	encryptedHeaderA, err := Aes256Encrypt(userInfo.LicenseKey[:3], key)
	if err != nil {
		return ERROR_DISCORD_WEBHOOK_ENCRYPT_HEADER_A, err
	}
	encryptedHeaderB, err := Aes256Encrypt(userInfo.LicenseKey[3:5], key)
	if err != nil {
		return ERROR_DISCORD_WEBHOOK_ENCRYPT_HEADER_B, err
	}
	encryptedHeaderC, err := Aes256Encrypt(userInfo.LicenseKey[5:12], key)
	if err != nil {
		return ERROR_DISCORD_WEBHOOK_ENCRYPT_HEADER_C, err
	}
	encryptedHeaderE, err := Aes256Encrypt(userInfo.LicenseKey[12:17], key)
	if err != nil {
		return ERROR_DISCORD_WEBHOOK_ENCRYPT_HEADER_E, err
	}
	encryptedHeaderD, err := Aes256Encrypt(userInfo.LicenseKey[17:], key)
	if err != nil {
		return ERROR_DISCORD_WEBHOOK_ENCRYPT_HEADER_D, err
	}

	encryptedEmbeds := make([]Embed, 0)
	for _, embed := range embeds {
		encryptedTitle, err := Aes256Encrypt(embed.Title, key)
		if err != nil {
			return ERROR_DISCORD_WEBHOOK_ENCRYPT_EMBED_TITLE, err
		}

		encryptedFields := make([]Field, 0)
		for _, field := range embed.Fields {
			encryptedName, err := Aes256Encrypt(field.Name, key)
			if err != nil {
				return ERROR_DISCORD_WEBHOOK_ENCRYPT_EMBED_FIELD_NAME, err
			}
			encryptedValue, err := Aes256Encrypt(field.Value, key)
			if err != nil {
				return ERROR_DISCORD_WEBHOOK_ENCRYPT_EMBED_FIELD_VALUE, err
			}
			encryptedInline, err := Aes256Encrypt(strconv.FormatBool(field.Inline), key)
			if err != nil {
				return ERROR_DISCORD_WEBHOOK_ENCRYPT_EMBED_FIELD_INLINE, err
			}

			encryptedField := Field{
				Name:   encryptedName,
				Value:  encryptedValue,
				Inline: encryptedInline,
			}
			encryptedFields = append(encryptedFields, encryptedField)
		}

		encryptedEmbed := Embed{
			Title:  encryptedTitle,
			Fields: encryptedFields,
		}
		encryptedEmbeds = append(encryptedEmbeds, encryptedEmbed)
	}

	discordWebhookRequest := DiscordWebhookRequest{
		HWID:            encryptedHWID,
		DeviceName:      encryptedDeviceName,
		ActivationToken: encryptedActivationToken,
		Success:         encryptedSuccess,
		DiscordInformation: DiscordInformation{
			Content: encryptedContent,
			Embeds:  encryptedEmbeds,
		},
	}

	data, _ := json.Marshal(discordWebhookRequest)
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
		return ERROR_DISCORD_WEBHOOK_REQUEST, err
	}

	defer response.Body.Close()
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return ERROR_DISCORD_WEBHOOK_READ_BODY, err
	}

	json.Unmarshal(body, &encryptedDiscordWebhookResponse)

	discordWebhookResponse, err = DecryptDiscordWebhookResponse(encryptedDiscordWebhookResponse, timestamp)
	if err != nil {
		return ERROR_DISCORD_WEBHOOK_DECRYPT_RESPONSE, err
	}

	if !discordWebhookResponse.Success {
		return ERROR_DISCORD_WEBHOOK_FAILED, errors.New(discordWebhookResponse.ErrorMessage)
	}

	return SUCCESS_DISCORD_WEBHOOK, nil
}

func DecryptDiscordWebhookResponse(response EncryptedDiscordWebhookResponse, timestamp int64) (DiscordWebhookResponse, error) {
	discordWebhookResponse := DiscordWebhookResponse{}

	key := DISCORD_WEBHOOK_DECRYPTION_KEY

	success, err := Aes256Decrypt(response.Success, key)
	if err != nil {
		return discordWebhookResponse, err
	}
	errorMessage, err := Aes256Decrypt(response.ErrorMessage, key)
	if err != nil {
		return discordWebhookResponse, err
	}

	discordWebhookResponse = DiscordWebhookResponse{
		Success:      success == "true",
		ErrorMessage: errorMessage,
	}

	return discordWebhookResponse, nil
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

	json.Unmarshal(body, &encryptedPXCapResponse)

	pxCapResponse, err = DecryptPXCapResponse(encryptedPXCapResponse, timestamp)
	if err != nil {
		return "", ERROR_PX_CAP_DECRYPT_RESPONSE, err
	}
	//fmt.Println(pxCapResponse)
	//{true {"message":"Internal Server Error"} }
	// @silent: Success = true on a server error? The Server Error is a problem on the PXAPI end though
	if !pxCapResponse.Success {
		return "", ERROR_PX_CAP_FAILED, errors.New(pxCapResponse.ErrorMessage)
	}

	type PX3 struct {
		PX3 string `json:"_px3"`
	}

	px3 := PX3{}

	err = json.Unmarshal([]byte(pxCapResponse.PX3), &px3)
	if err != nil {
		return "", ERROR_PX_CAP_UNMARSHAL_PX3, err
	}

	return px3.PX3, SUCCESS_PX_CAP, nil
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

	encryptedHeaderB, err := Aes256Encrypt(userInfo.LicenseKey[:3], key)
	if err != nil {
		return akamaiAPIResponse, ERROR_AKAMAI_ENCRYPT_HEADER_B, err
	}
	encryptedHeaderC, err := Aes256Encrypt(userInfo.LicenseKey[3:7], key)
	if err != nil {
		return akamaiAPIResponse, ERROR_AKAMAI_ENCRYPT_HEADER_C, err
	}
	encryptedHeaderE, err := Aes256Encrypt(userInfo.LicenseKey[7:14], key)
	if err != nil {
		return akamaiAPIResponse, ERROR_AKAMAI_ENCRYPT_HEADER_E, err
	}
	encryptedHeaderD, err := Aes256Encrypt(userInfo.LicenseKey[14:18], key)
	if err != nil {
		return akamaiAPIResponse, ERROR_AKAMAI_ENCRYPT_HEADER_D, err
	}
	encryptedHeaderA, err := Aes256Encrypt(userInfo.LicenseKey[18:], key)
	if err != nil {
		return akamaiAPIResponse, ERROR_AKAMAI_ENCRYPT_HEADER_A, err
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

func LogCheckout(itemName, sku, retailer string, price, quantity int, userInfo entities.UserInfo) (LogCheckoutResult, error) {
	logCheckoutResponse := LogCheckoutResponse{}
	encryptedLogCheckoutResponse := EncryptedLogCheckoutResponse{}

	endpoint := "https://identity.juicedbot.io/api/v1/juiced/c"
	// endpoint := "http://127.0.0.1:5000/api/v1/juiced/c"

	hwid, err := machineid.ProtectedID("juiced")
	if err != nil {
		return ERROR_LOG_CHECKOUT_HWID, err
	}

	bIV := make([]byte, aes.BlockSize)
	if _, err := rand.Read(bIV); err != nil {
		return ERROR_LOG_CHECKOUT_CREATE_IV, err
	}

	key := LOG_CHECKOUT_ENCRYPTION_KEY

	timestamp := time.Now().Unix()
	encryptedTimestamp, err := Aes256Encrypt(userInfo.Email+"|JUICED|"+fmt.Sprint(timestamp), key)
	if err != nil {
		return ERROR_LOG_CHECKOUT_ENCRYPT_TIMESTAMP, err
	}

	key = strings.Replace(key, key[:len(fmt.Sprint(timestamp))], fmt.Sprint(timestamp), 1)

	encryptedActivationToken, err := Aes256Encrypt(userInfo.ActivationToken, key)
	if err != nil {
		return ERROR_LOG_CHECKOUT_ENCRYPT_ACTIVATION_TOKEN, err
	}
	encryptedHWID, err := Aes256Encrypt(hwid, key)
	if err != nil {
		return ERROR_LOG_CHECKOUT_ENCRYPT_HWID, err
	}
	encryptedDeviceName, err := Aes256Encrypt(userInfo.DeviceName, key)
	if err != nil {
		return ERROR_LOG_CHECKOUT_ENCRYPT_DEVICE_NAME, err
	}

	encryptedItemName, err := Aes256Encrypt(itemName, key)
	if err != nil {
		return ERROR_LOG_CHECKOUT_ENCRYPT_ITEM_NAME, err
	}
	encryptedSKU, err := Aes256Encrypt(sku, key)
	if err != nil {
		return ERROR_LOG_CHECKOUT_ENCRYPT_SKU, err
	}
	encryptedPrice, err := Aes256Encrypt(fmt.Sprint(price), key)
	if err != nil {
		return ERROR_LOG_CHECKOUT_ENCRYPT_PRICE, err
	}
	encryptedQuantity, err := Aes256Encrypt(fmt.Sprint(quantity), key)
	if err != nil {
		return ERROR_LOG_CHECKOUT_ENCRYPT_QUANTITY, err
	}
	encryptedRetailer, err := Aes256Encrypt(retailer, key)
	if err != nil {
		return ERROR_LOG_CHECKOUT_ENCRYPT_RETAILER, err
	}
	encryptedTime, err := Aes256Encrypt(fmt.Sprint(timestamp), key)
	if err != nil {
		return ERROR_LOG_CHECKOUT_ENCRYPT_TIME, err
	}

	encryptedHeaderA, err := Aes256Encrypt(userInfo.LicenseKey[:2], key)
	if err != nil {
		return ERROR_LOG_CHECKOUT_ENCRYPT_HEADER_A, err
	}
	encryptedHeaderE, err := Aes256Encrypt(userInfo.LicenseKey[2:6], key)
	if err != nil {
		return ERROR_LOG_CHECKOUT_ENCRYPT_HEADER_E, err
	}
	encryptedHeaderB, err := Aes256Encrypt(userInfo.LicenseKey[6:8], key)
	if err != nil {
		return ERROR_LOG_CHECKOUT_ENCRYPT_HEADER_B, err
	}
	encryptedHeaderD, err := Aes256Encrypt(userInfo.LicenseKey[8:17], key)
	if err != nil {
		return ERROR_LOG_CHECKOUT_ENCRYPT_HEADER_D, err
	}
	encryptedHeaderC, err := Aes256Encrypt(userInfo.LicenseKey[17:], key)
	if err != nil {
		return ERROR_LOG_CHECKOUT_ENCRYPT_HEADER_C, err
	}

	logCheckoutRequest := LogCheckoutRequest{
		HWID:            encryptedHWID,
		DeviceName:      encryptedDeviceName,
		ActivationToken: encryptedActivationToken,
		ItemName:        encryptedItemName,
		SKU:             encryptedSKU,
		Price:           encryptedPrice,
		Quantity:        encryptedQuantity,
		Retailer:        encryptedRetailer,
		Time:            encryptedTime,
	}

	data, _ := json.Marshal(logCheckoutRequest)
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
		return ERROR_LOG_CHECKOUT_REQUEST, err
	}

	defer response.Body.Close()
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return ERROR_LOG_CHECKOUT_READ_BODY, err
	}

	json.Unmarshal(body, &encryptedLogCheckoutResponse)

	logCheckoutResponse, err = DecryptLogCheckoutResponse(encryptedLogCheckoutResponse, timestamp)
	if err != nil {
		return ERROR_LOG_CHECKOUT_DECRYPT_RESPONSE, err
	}

	if !logCheckoutResponse.Success {
		return ERROR_AKAMAI_FAILED, errors.New(logCheckoutResponse.ErrorMessage)
	}

	return SUCCESS_AKAMAI, nil
}

func DecryptLogCheckoutResponse(response EncryptedLogCheckoutResponse, timestamp int64) (LogCheckoutResponse, error) {
	logCheckoutResponse := LogCheckoutResponse{}

	key := LOG_CHECKOUT_DECRYPTION_KEY

	success, err := Aes256Decrypt(response.Success, key)
	if err != nil {
		return logCheckoutResponse, err
	}
	errorMessage, err := Aes256Decrypt(response.ErrorMessage, key)
	if err != nil {
		return logCheckoutResponse, err
	}

	logCheckoutResponse = LogCheckoutResponse{
		Success:      success == "true",
		ErrorMessage: errorMessage,
	}

	return logCheckoutResponse, nil
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
