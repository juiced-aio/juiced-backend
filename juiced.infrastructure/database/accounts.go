package database

import (
	"encoding/json"
	"fmt"
	"time"

	"backend.juicedbot.io/juiced.client/http"
	"backend.juicedbot.io/juiced.infrastructure/entities"
	"backend.juicedbot.io/juiced.infrastructure/enums"
	"backend.juicedbot.io/juiced.infrastructure/util"
)

func GetAllAccounts() ([]entities.Account, error) {
	accounts := []entities.Account{}
	if database == nil {
		return accounts, &DatabaseNotInitializedError{}
	}

	rows, err := database.Queryx("SELECT * FROM accounts")
	if err != nil {
		return accounts, err
	}

	defer rows.Close()
	for rows.Next() {
		account := entities.Account{}
		err = rows.StructScan(&account)
		if err != nil {
			return accounts, err
		}

		var encryptedEmail string
		var encryptedPassword string
		decryptedEmail, err := util.Aes256Decrypt(account.Email, enums.UserKey)
		if err == nil {
			account.Email = decryptedEmail
		} else {
			encryptedEmail, err = util.Aes256Encrypt(account.Email, enums.UserKey)
			if err != nil {
				return accounts, err
			}
		}

		decryptedPassword, err := util.Aes256Decrypt(account.Password, enums.UserKey)
		if err == nil {
			account.Password = decryptedPassword
		} else {
			encryptedPassword, err = util.Aes256Encrypt(account.Password, enums.UserKey)
			if err != nil {
				return accounts, err
			}
		}

		if encryptedEmail != "" {
			go func() {
				for {
					_, err = database.Exec(fmt.Sprintf(`UPDATE accounts SET email = "%v" WHERE ID = "%v"`, encryptedEmail, account.ID))
					if err != nil {
						time.Sleep(1 * time.Second)
						continue
					} else {
						break
					}
				}
			}()
		}

		if encryptedPassword != "" {
			go func() {
				for {
					_, err = database.Exec(fmt.Sprintf(`UPDATE accounts SET password = "%v" WHERE ID = "%v"`, encryptedPassword, account.ID))
					if err != nil {
						time.Sleep(1 * time.Second)
						continue
					} else {
						break
					}
				}
			}()
		}

		cookies := []http.Cookie{}
		if account.CookiesSerialized != "" {
			err = json.Unmarshal([]byte(account.CookiesSerialized), &cookies)
			if err == nil {
				cookiePtrs := []*http.Cookie{}
				for _, cookie := range cookies {
					cookiePtrs = append(cookiePtrs, &cookie)
				}
				account.Cookies = cookiePtrs
			}
		}

		accounts = append(accounts, account)
	}

	return accounts, nil
}

func GetAccount(ID string) (entities.Account, error) {
	account := entities.Account{}
	if database == nil {
		return account, &DatabaseNotInitializedError{}
	}

	statement, err := database.Preparex("SELECT * FROM accounts WHERE ID = @p1")
	if err != nil {
		return account, err
	}

	rows, err := statement.Queryx(ID)
	if err != nil {
		return account, err
	}

	defer rows.Close()
	for rows.Next() {
		err = rows.StructScan(&account)
		if err != nil {
			return account, err
		}

		var encryptedEmail string
		var encryptedPassword string
		decryptedEmail, err := util.Aes256Decrypt(account.Email, enums.UserKey)
		if err == nil {
			account.Email = decryptedEmail
		} else {
			encryptedEmail, err = util.Aes256Encrypt(account.Email, enums.UserKey)
			if err != nil {
				return account, err
			}
		}

		decryptedPassword, err := util.Aes256Decrypt(account.Password, enums.UserKey)
		if err == nil {
			account.Password = decryptedPassword
		} else {
			encryptedPassword, err = util.Aes256Encrypt(account.Password, enums.UserKey)
			if err != nil {
				return account, err
			}
		}

		if encryptedEmail != "" {
			go func() {
				for {
					_, err = database.Exec(fmt.Sprintf(`UPDATE accounts SET email = "%v" WHERE ID = "%v"`, encryptedEmail, account.ID))
					if err != nil {
						time.Sleep(1 * time.Second)
						continue
					} else {
						break
					}
				}
			}()
		}

		if encryptedPassword != "" {
			go func() {
				for {
					_, err = database.Exec(fmt.Sprintf(`UPDATE accounts SET password = "%v" WHERE ID = "%v"`, encryptedPassword, account.ID))
					if err != nil {
						time.Sleep(1 * time.Second)
						continue
					} else {
						break
					}
				}
			}()
		}
	}

	cookies := []http.Cookie{}
	if account.CookiesSerialized != "" {
		err = json.Unmarshal([]byte(account.CookiesSerialized), &cookies)
		if err == nil {
			cookiePtrs := []*http.Cookie{}
			for _, cookie := range cookies {
				cookiePtrs = append(cookiePtrs, &cookie)
			}
			account.Cookies = cookiePtrs
		}
	}

	return account, nil
}

func CreateAccount(account entities.Account) error {
	if database == nil {
		return &DatabaseNotInitializedError{}
	}

	encryptedEmail, err := util.Aes256Encrypt(account.Email, enums.UserKey)
	if err != nil {
		return err
	}

	encryptedPassword, err := util.Aes256Encrypt(account.Password, enums.UserKey)
	if err != nil {
		return err
	}

	cookiesSerialized, err := json.Marshal(account.Cookies)
	if err != nil {
		return err
	}

	statement, err := database.Preparex(`INSERT INTO accounts (ID, retailer, email, password, cookiesSerialized, creationDate) VALUES (?, ?, ?, ?, ?, ?)`)
	if err != nil {
		return err
	}

	_, err = statement.Exec(account.ID, account.Retailer, encryptedEmail, encryptedPassword, cookiesSerialized, account.CreationDate)

	return err
}

func UpdateAccount(ID string, newAccount entities.Account) error {
	err := RemoveAccount(ID)
	if err != nil {
		return err
	}
	return CreateAccount(newAccount)
}

func RemoveAccount(ID string) error {
	if database == nil {
		return &DatabaseNotInitializedError{}
	}
	statement, err := database.Preparex(`DELETE FROM accounts WHERE ID = @p1`)
	if err != nil {
		return err
	}
	_, err = statement.Exec(ID)

	return err
}
