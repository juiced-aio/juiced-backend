package util

// AUTHENTICATION_ENCRYPTION_KEY is the encryption key used to encrypt the fields and values for the request to POST /juiced/a
const AUTHENTICATION_ENCRYPTION_KEY = "4CD9DA145AD79FC817DCD3A63D50285F"

// AUTHENTICATION_DECRYPTION_KEY is the encryption key used to decrypt the fields and values for the response from POST /juiced/a
const AUTHENTICATION_DECRYPTION_KEY = "42618E27E9EB15456AB219A8BD336C5F"

type AuthenticationResult = int

const (
	ERROR_AUTHENTICATE_HWID AuthenticationResult = iota
	ERROR_AUTHENTICATE_CREATE_IV
	ERROR_AUTHENTICATE_ENCRYPT_TIMESTAMP
	ERROR_AUTHENTICATE_ENCRYPT_ACTIVATION_TOKEN
	ERROR_AUTHENTICATE_ENCRYPT_HWID
	ERROR_AUTHENTICATE_ENCRYPT_DEVICE_NAME
	ERROR_AUTHENTICATE_ENCRYPT_HEADER_A
	ERROR_AUTHENTICATE_ENCRYPT_HEADER_B
	ERROR_AUTHENTICATE_ENCRYPT_HEADER_C
	ERROR_AUTHENTICATE_ENCRYPT_HEADER_D
	ERROR_AUTHENTICATE_ENCRYPT_HEADER_E
	ERROR_AUTHENTICATE_REQUEST
	ERROR_AUTHENTICATE_READ_BODY
	ERROR_AUTHENTICATE_DECRYPT_RESPONSE
	ERROR_AUTHENTICATE_TOKEN_EXPIRED
	ERROR_AUTHENTICATE_FAILED
	SUCCESS_AUTHENTICATE
)

// REFRESH_ENCRYPTION_KEY is the encryption key used to encrypt the fields and values for the request to POST /juiced/r
const REFRESH_ENCRYPTION_KEY = "D4C891366401815C008987157D364BF2"

// REFRESH_DECRYPTION_KEY is the encryption key used to decrypt the fields and values for the response from POST /juiced/r
const REFRESH_DECRYPTION_KEY = "CB13AA963AA17352E765AAD59C0B3A5B"

type RefreshResult = int

const (
	ERROR_REFRESH_HWID RefreshResult = iota
	ERROR_REFRESH_CREATE_IV
	ERROR_REFRESH_ENCRYPT_TIMESTAMP
	ERROR_REFRESH_ENCRYPT_ACTIVATION_TOKEN
	ERROR_REFRESH_ENCRYPT_REFRESH_TOKEN
	ERROR_REFRESH_ENCRYPT_HWID
	ERROR_REFRESH_ENCRYPT_DEVICE_NAME
	ERROR_REFRESH_ENCRYPT_HEADER_A
	ERROR_REFRESH_ENCRYPT_HEADER_B
	ERROR_REFRESH_ENCRYPT_HEADER_C
	ERROR_REFRESH_ENCRYPT_HEADER_D
	ERROR_REFRESH_ENCRYPT_HEADER_E
	ERROR_REFRESH_REQUEST
	ERROR_REFRESH_READ_BODY
	ERROR_REFRESH_DECRYPT_RESPONSE
	ERROR_REFRESH_FAILED
	SUCCESS_REFRESH_ERROR_SET_USER_INFO
	SUCCESS_REFRESH
)

// PX_ENCRYPTION_KEY is the encryption key used to encrypt the fields and values for the request to POST /juiced/p
const PX_ENCRYPTION_KEY = "241406D242D9C1AA458C8271D1347E30"

// PX_DECRYPTION_KEY is the encryption key used to decrypt the fields and values for the response from POST /juiced/p
const PX_DECRYPTION_KEY = "D32D87264D2A95069A0B58C3CD682BBE"

type PXResult = int

const (
	ERROR_PX_HWID PXResult = iota
	ERROR_PX_CREATE_IV
	ERROR_PX_ENCRYPT_TIMESTAMP
	ERROR_PX_ENCRYPT_ACTIVATION_TOKEN
	ERROR_PX_ENCRYPT_HWID
	ERROR_PX_ENCRYPT_DEVICE_NAME
	ERROR_PX_ENCRYPT_HEADER_A
	ERROR_PX_ENCRYPT_HEADER_B
	ERROR_PX_ENCRYPT_HEADER_C
	ERROR_PX_ENCRYPT_HEADER_D
	ERROR_PX_ENCRYPT_HEADER_E
	ERROR_PX_REQUEST
	ERROR_PX_READ_BODY
	ERROR_PX_DECRYPT_RESPONSE
	ERROR_PX_FAILED
	SUCCESS_PX_ERROR_SET_USER_INFO
	SUCCESS_PX
)

// PXCAP_ENCRYPTION_KEY is the encryption key used to encrypt the fields and values for the request to POST /juiced/pc
const PXCAP_ENCRYPTION_KEY = "663947D3BE2963ED5CA2C3227246C72B"

// PXCAP_DECRYPTION_KEY is the encryption key used to decrypt the fields and values for the response from POST /juiced/pc
const PXCAP_DECRYPTION_KEY = "3CCA9F922797192F41BE09A6C9955450"

type PXCapResult = int

const (
	ERROR_PX_CAP_HWID PXCapResult = iota
	ERROR_PX_CAP_CREATE_IV
	ERROR_PX_CAP_ENCRYPT_TIMESTAMP
	ERROR_PX_CAP_ENCRYPT_ACTIVATION_TOKEN
	ERROR_PX_CAP_ENCRYPT_HWID
	ERROR_PX_CAP_ENCRYPT_DEVICE_NAME
	ERROR_PX_CAP_ENCRYPT_HEADER_A
	ERROR_PX_CAP_ENCRYPT_HEADER_B
	ERROR_PX_CAP_ENCRYPT_HEADER_C
	ERROR_PX_CAP_ENCRYPT_HEADER_D
	ERROR_PX_CAP_ENCRYPT_HEADER_E
	ERROR_PX_CAP_REQUEST
	ERROR_PX_CAP_READ_BODY
	ERROR_PX_CAP_DECRYPT_RESPONSE
	ERROR_PX_CAP_FAILED
	ERROR_PX_CAP_UNMARSHAL_PX3
	SUCCESS_PX_CAP_ERROR_SET_USER_INFO
	SUCCESS_PX_CAP
)

// AKAMAI_ENCRYPTION_KEY is the encryption key used to encrypt the fields and values for the request to POST /juiced/ak
const AKAMAI_ENCRYPTION_KEY = "ED822E2446C1DFCBBF401EBFE7DDEB45"

// AKAMAI_DECRYPTION_KEY is the encryption key used to decrypt the fields and values for the response from POST /juiced/ak
const AKAMAI_DECRYPTION_KEY = "DB17E0E733DBC50F55ECA71311F0E562"

type AkamaiResult = int

const (
	ERROR_AKAMAI_HWID AkamaiResult = iota
	ERROR_AKAMAI_CREATE_IV
	ERROR_AKAMAI_ENCRYPT_TIMESTAMP
	ERROR_AKAMAI_ENCRYPT_ACTIVATION_TOKEN
	ERROR_AKAMAI_ENCRYPT_HWID
	ERROR_AKAMAI_ENCRYPT_DEVICE_NAME
	ERROR_AKAMAI_ENCRYPT_PAGE_URL
	ERROR_AKAMAI_ENCRYPT_SKIP_KACT
	ERROR_AKAMAI_ENCRYPT_SKIP_MACT
	ERROR_AKAMAI_ENCRYPT_ON_BLUR
	ERROR_AKAMAI_ENCRYPT_ON_FOCUS
	ERROR_AKAMAI_ENCRYPT_ABCK
	ERROR_AKAMAI_ENCRYPT_SENSOR_DATA_LINK
	ERROR_AKAMAI_ENCRYPT_VER
	ERROR_AKAMAI_ENCRYPT_FIRST_POST
	ERROR_AKAMAI_ENCRYPT_PIXEL_ID
	ERROR_AKAMAI_ENCRYPT_PIXEL_G
	ERROR_AKAMAI_ENCRYPT_JSON
	ERROR_AKAMAI_ENCRYPT_BASE_URL
	ERROR_AKAMAI_ENCRYPT_USER_AGENT
	ERROR_AKAMAI_ENCRYPT_COOKIE
	ERROR_AKAMAI_ENCRYPT_POST_INDX
	ERROR_AKAMAI_ENCRYPT_SAVED_D3
	ERROR_AKAMAI_ENCRYPT_SAVED_START_TS
	ERROR_AKAMAI_ENCRYPT_DEVICE_NUM
	ERROR_AKAMAI_ENCRYPT_HEADER_A
	ERROR_AKAMAI_ENCRYPT_HEADER_B
	ERROR_AKAMAI_ENCRYPT_HEADER_C
	ERROR_AKAMAI_ENCRYPT_HEADER_D
	ERROR_AKAMAI_ENCRYPT_HEADER_E
	ERROR_AKAMAI_REQUEST
	ERROR_AKAMAI_READ_BODY
	ERROR_AKAMAI_DECRYPT_RESPONSE
	ERROR_AKAMAI_FAILED
	SUCCESS_AKAMAI_ERROR_SET_USER_INFO
	SUCCESS_AKAMAI
)

// DISCORD_WEBHOOK_ENCRYPTION_KEY is the encryption key used to encrypt the fields and values for the request to POST /juiced/dw
const DISCORD_WEBHOOK_ENCRYPTION_KEY = "DC1A492D4F524CBDAA387CF465D6A4D0"

// DISCORD_WEBHOOK_DECRYPTION_KEY is the encryption key used to decrypt the fields and values for the response from POST /juiced/dw
const DISCORD_WEBHOOK_DECRYPTION_KEY = "1731639116D23C28DB26B5440422E2AC"

type DiscordWebhookResult = int

const (
	ERROR_DISCORD_WEBHOOK_HWID DiscordWebhookResult = iota
	ERROR_DISCORD_WEBHOOK_CREATE_IV
	ERROR_DISCORD_WEBHOOK_ENCRYPT_TIMESTAMP
	ERROR_DISCORD_WEBHOOK_ENCRYPT_ACTIVATION_TOKEN
	ERROR_DISCORD_WEBHOOK_ENCRYPT_HWID
	ERROR_DISCORD_WEBHOOK_ENCRYPT_DEVICE_NAME
	ERROR_DISCORD_WEBHOOK_ENCRYPT_SUCCESS
	ERROR_DISCORD_WEBHOOK_ENCRYPT_CONTENT
	ERROR_DISCORD_WEBHOOK_ENCRYPT_EMBED_TITLE
	ERROR_DISCORD_WEBHOOK_ENCRYPT_EMBED_FIELD_NAME
	ERROR_DISCORD_WEBHOOK_ENCRYPT_EMBED_FIELD_VALUE
	ERROR_DISCORD_WEBHOOK_ENCRYPT_EMBED_FIELD_INLINE
	ERROR_DISCORD_WEBHOOK_ENCRYPT_HEADER_A
	ERROR_DISCORD_WEBHOOK_ENCRYPT_HEADER_B
	ERROR_DISCORD_WEBHOOK_ENCRYPT_HEADER_C
	ERROR_DISCORD_WEBHOOK_ENCRYPT_HEADER_D
	ERROR_DISCORD_WEBHOOK_ENCRYPT_HEADER_E
	ERROR_DISCORD_WEBHOOK_REQUEST
	ERROR_DISCORD_WEBHOOK_READ_BODY
	ERROR_DISCORD_WEBHOOK_DECRYPT_RESPONSE
	ERROR_DISCORD_WEBHOOK_FAILED
	SUCCESS_DISCORD_WEBHOOK_ERROR_SET_USER_INFO
	SUCCESS_DISCORD_WEBHOOK
)

// LOG_CHECKOUT_ENCRYPTION_KEY is the encryption key used to encrypt the fields and values for the request to POST /juiced/c
const LOG_CHECKOUT_ENCRYPTION_KEY = "BCBD40707BF64211F5535405E52644B3"

// LOG_CHECKOUT_DECRYPTION_KEY is the encryption key used to decrypt the fields and values for the response from POST /juiced/c
const LOG_CHECKOUT_DECRYPTION_KEY = "8C6D57F6174F6AC45B56E88313F41C2D"

type LogCheckoutResult = int

const (
	ERROR_LOG_CHECKOUT_HWID LogCheckoutResult = iota
	ERROR_LOG_CHECKOUT_CREATE_IV
	ERROR_LOG_CHECKOUT_ENCRYPT_TIMESTAMP
	ERROR_LOG_CHECKOUT_ENCRYPT_ACTIVATION_TOKEN
	ERROR_LOG_CHECKOUT_ENCRYPT_HWID
	ERROR_LOG_CHECKOUT_ENCRYPT_DEVICE_NAME
	ERROR_LOG_CHECKOUT_ENCRYPT_ITEM_NAME
	ERROR_LOG_CHECKOUT_ENCRYPT_SKU
	ERROR_LOG_CHECKOUT_ENCRYPT_PRICE
	ERROR_LOG_CHECKOUT_ENCRYPT_QUANTITY
	ERROR_LOG_CHECKOUT_ENCRYPT_RETAILER
	ERROR_LOG_CHECKOUT_ENCRYPT_TIME
	ERROR_LOG_CHECKOUT_ENCRYPT_HEADER_A
	ERROR_LOG_CHECKOUT_ENCRYPT_HEADER_B
	ERROR_LOG_CHECKOUT_ENCRYPT_HEADER_C
	ERROR_LOG_CHECKOUT_ENCRYPT_HEADER_D
	ERROR_LOG_CHECKOUT_ENCRYPT_HEADER_E
	ERROR_LOG_CHECKOUT_REQUEST
	ERROR_LOG_CHECKOUT_READ_BODY
	ERROR_LOG_CHECKOUT_DECRYPT_RESPONSE
	ERROR_LOG_CHECKOUT_FAILED
	SUCCESS_LOG_CHECKOUT_ERROR_SET_USER_INFO
	SUCCESS_LOG_CHECKOUT
)

// GET_ENCRYPTION_KEY_ENCRYPTION_KEY is the encryption key used to encrypt the fields and values for the request to POST /juiced/e
const GET_ENCRYPTION_KEY_ENCRYPTION_KEY = "FB6D08F24CB53AE2C19D9458A74FE26E"

// GET_ENCRYPTION_KEY_DECRYPTION_KEY is the encryption key used to decrypt the fields and values for the response from POST /juiced/e
const GET_ENCRYPTION_KEY_DECRYPTION_KEY = "557F0C6239CD24122B9D7033576CCF56"

type GetEncryptionKeyResult = int

const (
	ERROR_GET_ENCRYPTION_KEY_HWID GetEncryptionKeyResult = iota
	ERROR_GET_ENCRYPTION_KEY_CREATE_IV
	ERROR_GET_ENCRYPTION_KEY_ENCRYPT_TIMESTAMP
	ERROR_GET_ENCRYPTION_KEY_ENCRYPT_ACTIVATION_TOKEN
	ERROR_GET_ENCRYPTION_KEY_ENCRYPT_HWID
	ERROR_GET_ENCRYPTION_KEY_ENCRYPT_DEVICE_NAME
	ERROR_GET_ENCRYPTION_KEY_ENCRYPT_HEADER_A
	ERROR_GET_ENCRYPTION_KEY_ENCRYPT_HEADER_B
	ERROR_GET_ENCRYPTION_KEY_ENCRYPT_HEADER_C
	ERROR_GET_ENCRYPTION_KEY_ENCRYPT_HEADER_D
	ERROR_GET_ENCRYPTION_KEY_ENCRYPT_HEADER_E
	ERROR_GET_ENCRYPTION_KEY_REQUEST
	ERROR_GET_ENCRYPTION_KEY_READ_BODY
	ERROR_GET_ENCRYPTION_KEY_UNMARSHAL_BODY
	ERROR_GET_ENCRYPTION_KEY_DECRYPT_RESPONSE
	ERROR_GET_ENCRYPTION_KEY_TOKEN_EXPIRED
	ERROR_GET_ENCRYPTION_KEY_FAILED
	SUCCESS_GET_ENCRYPTION_KEY
)

// MAX_RETRIES is the number of times the app will retry a heartbeat request before closing
const MAX_RETRIES = 5

type AuthErrorCode = int

const (
	SUCCESS AuthErrorCode = iota
	NO_STORED_INFO
	ERROR_CONNECTING_TO_DATABASE
	ERROR_AUTHENTICATING_EXISTING_INFO
	ERROR_READING_REQUEST_BODY
	ERROR_BEFORE_REQUEST
	ERROR_DURING_REQUEST
	ERROR_AFTER_REQUEST
)
