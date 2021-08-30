package cybersource

type Encrypt struct {
	Flx struct {
		Jwk struct {
			Kty string
			E   string
			Use string
			N   string
			Kid string
		}
	}
}

type RSA struct {
	Kty string `json:"kty"`
	E   string `json:"e"`
	Use string `json:"use"`
	Kid string `json:"kid"`
	N   string `json:"n"`
}

type Header__ struct {
	Kid string `json:"kid"`
	Jwk RSA    `json:"jwk"`
}

type EncryptedObject struct {
	Context string `json:"context"`
	Index   int    `json:"index"`
	Data    Card   `json:"data"`
}

type PaymentToken struct {
	Jti string `json:"jti"`
}

type Card struct {
	SecurityCode string `json:"securityCode"`
	Number       string `json:"number"`
	Type         string `json:"type"`
	ExpMonth     string `json:"expMonth"`
	ExpYear      string `json:"expYear"`
}
