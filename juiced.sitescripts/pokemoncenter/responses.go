package pokemoncenter

type MonitorResponse struct {
	Props struct {
		IsServer     bool
		InitialState struct {
			Product struct {
				AddToCartForm string
				Availability  string
				Images        []struct {
					High     string
					Original string
				}
				ListPrice struct {
					Amount float64
				}
				Name string
			}
		}
		InitialProps struct {
		}
	}
}

type AccessToken struct {
	AccessToken string
}

type AddToCartResponse struct {
	Type     string `json:"type"`
	Quantity int    `json:"quantity"`
}

type SubmitPaymentResponse struct {
	Self struct {
		Uri string
	}
}

type PaymentKeyResponse struct {
	KeyId string `json:"keyId"`
}

type LoginResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   string `json:"expires_in"`
	Scope       string `json:"scope"`
	Role        string `json:"role"`
	Roles       string `json:"roles"`
	FamilyName  string `json:"familyName"`
	GivenName   string `json:"givenName"`
}
