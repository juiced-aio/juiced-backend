package entities

import (
	"backend.juicedbot.io/juiced.client/http"
	"backend.juicedbot.io/juiced.infrastructure/enums"
)

type Account struct {
	ID                string              `json:"ID" db:"ID"`
	Retailer          enums.Retailer      `json:"retailer" db:"retailer"`
	Email             string              `json:"email" db:"email"`
	Password          string              `json:"password" db:"password"`
	CookiesSerialized string              `json:"-" db:"cookiesSerialized"`
	Cookies           []*http.Cookie      `json:"-"`
	Status            enums.AccountStatus `json:"status"`
	IsTemp            bool                `json:"-"`
	CreationDate      int64               `json:"creationDate" db:"creationDate"`
}
