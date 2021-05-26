package util

// AUTHENTICATION_ENCRYPTION_KEY is the encryption key used to encrypt the fields and values for the request to POST /juiced/a
const AUTHENTICATION_ENCRYPTION_KEY = "4CD9DA145AD79FC817DCD3A63D50285F"

// AUTHENTICATION_DECRYPTION_KEY is the encryption key used to decrypt the fields and values for the response from POST /juiced/a
const AUTHENTICATION_DECRYPTION_KEY = "42618E27E9EB15456AB219A8BD336C5F"

// ACTIVATION_ENCRYPTION_KEY is the encryption key used to encrypt the fields and values for the request to POST /juiced/ac
const ACTIVATION_ENCRYPTION_KEY = "5EAD4B42F2E35643BC5A1C13844AD3BA"

// ACTIVATION_DECRYPTION_KEY is the encryption key used to decrypt the fields and values for the response from POST /juiced/ac
const ACTIVATION_DECRYPTION_KEY = "395FA0BD77B72E3FA0D52BCFB89799FA"

// DEACTIVATION_ENCRYPTION_KEY is the encryption key used to encrypt the fields and values for the request to POST /juiced/d
const DEACTIVATION_ENCRYPTION_KEY = "52DB549C1EF7EEBEE55B855C114CD1SM"

// DEACTIVATION_DECRYPTION_KEY is the encryption key used to decrypt the fields and values for the response from POST /juiced/d
const DEACTIVATION_DECRYPTION_KEY = "D34714CCF7D2E613CECA19FA28E9DSLK"

// REFRESH_ENCRYPTION_KEY is the encryption key used to encrypt the fields and values for the request to POST /juiced/r
const REFRESH_ENCRYPTION_KEY = "D4C891366401815C008987157D364BF2"

// REFRESH_DECRYPTION_KEY is the encryption key used to decrypt the fields and values for the response from POST /juiced/r
const REFRESH_DECRYPTION_KEY = "CB13AA963AA17352E765AAD59C0B3A5B"

// DOWNLOAD_ENCRYPTION_KEY is the encryption key used to encrypt the fields and values for the response from POST /juiced/m and POST /juiced/w
const DOWNLOAD_ENCRYPTION_KEY = "C943BF22740C933456A563EA9544DEB2"

// PX_ENCRYPTION_KEY is the encryption key used to encrypt the fields and values for the request to POST /juiced/p
const PX_ENCRYPTION_KEY = "241406D242D9C1AA458C8271D1347E30"

// PX_DECRYPTION_KEY is the encryption key used to decrypt the fields and values for the response from POST /juiced/p
const PX_DECRYPTION_KEY = "D32D87264D2A95069A0B58C3CD682BBE"

// PXCAP_ENCRYPTION_KEY is the encryption key used to encrypt the fields and values for the request to POST /juiced/pc
const PXCAP_ENCRYPTION_KEY = "663947D3BE2963ED5CA2C3227246C72B"

// PXCAP_DECRYPTION_KEY is the encryption key used to decrypt the fields and values for the response from POST /juiced/pc
const PXCAP_DECRYPTION_KEY = "3CCA9F922797192F41BE09A6C9955450"

// AKAMAI_ENCRYPTION_KEY is the encryption key used to encrypt the fields and values for the request to POST /juiced/ak
const AKAMAI_ENCRYPTION_KEY = "ED822E2446C1DFCBBF401EBFE7DDEB45"

// AKAMAI_DECRYPTION_KEY is the encryption key used to decrypt the fields and values for the response from POST /juiced/ak
const AKAMAI_DECRYPTION_KEY = "DB17E0E733DBC50F55ECA71311F0E562"

// DISCORD_WEBHOOK_ENCRYPTION_KEY is the encryption key used to encrypt the fields and values for the request to POST /juiced/dw
const DISCORD_WEBHOOK_ENCRYPTION_KEY = "DC1A492D4F524CBDAA387CF465D6A4D0"

// DISCORD_WEBHOOK_DECRYPTION_KEY is the encryption key used to decrypt the fields and values for the response from POST /juiced/dw
const DISCORD_WEBHOOK_DECRYPTION_KEY = "1731639116D23C28DB26B5440422E2AC"

// MAX_RETRIES is the number of times the app will retry a heartbeat request before closing
const MAX_RETRIES = 5
