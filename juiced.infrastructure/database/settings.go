package database

import (
	"fmt"
	"sort"
	"time"

	"backend.juicedbot.io/juiced.infrastructure/common"
	"backend.juicedbot.io/juiced.infrastructure/entities"
	"backend.juicedbot.io/juiced.infrastructure/enums"
)

// GetSettings returns the settings object from the database
func GetSettings() (entities.Settings, error) {
	settings := entities.Settings{}
	if database == nil {
		return settings, &DatabaseNotInitializedError{}
	}

	// Might want to add "WHERE id = 0" to the query
	rows, err := database.Queryx("SELECT * FROM settings")
	if err != nil {
		return settings, err
	}

	defer rows.Close()
	for rows.Next() {
		err = rows.StructScan(&settings)
		if err != nil {
			return settings, err
		}
	}
	settings.Accounts, err = GetAccounts()
	return settings, err
}

// GetAccounts returns a list of accounts from the database
func GetAccounts() ([]entities.Account, error) {
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
		decryptedEmail, err := common.Aes256Decrypt(account.Email, enums.UserKey)
		if err == nil {
			account.Email = decryptedEmail
		} else {
			encryptedEmail, err = common.Aes256Encrypt(account.Email, enums.UserKey)
			if err != nil {
				return accounts, err
			}
		}

		decryptedPassword, err := common.Aes256Decrypt(account.Password, enums.UserKey)
		if err == nil {
			account.Password = decryptedPassword
		} else {
			encryptedPassword, err = common.Aes256Encrypt(account.Password, enums.UserKey)
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

		accounts = append(accounts, account)
	}

	sort.SliceStable(accounts, func(i, j int) bool {
		return accounts[i].CreationDate < accounts[j].CreationDate
	})

	return accounts, nil
}

// GetAccount returns an account from the database
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
		decryptedEmail, err := common.Aes256Decrypt(account.Email, enums.UserKey)
		if err == nil {
			account.Email = decryptedEmail
		} else {
			encryptedEmail, err = common.Aes256Encrypt(account.Email, enums.UserKey)
			if err != nil {
				return account, err
			}
		}

		decryptedPassword, err := common.Aes256Decrypt(account.Password, enums.UserKey)
		if err == nil {
			account.Password = decryptedPassword
		} else {
			encryptedPassword, err = common.Aes256Encrypt(account.Password, enums.UserKey)
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

	return account, nil
}

// UpdateSettings updates the Settings object in the database
func UpdateSettings(settings entities.Settings) error {
	if database == nil {
		return &DatabaseNotInitializedError{}
	}

	_, err := database.Exec("DELETE FROM settings")
	if err != nil {
		return err
	}

	statement, err := database.Preparex(`INSERT INTO settings (id, successDiscordWebhook, failureDiscordWebhook, twoCaptchaAPIKey, antiCaptchaAPIKey, capMonsterAPIKey, aycdAccessToken, aycdAPIKey, darkMode, useAnimations) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`)
	if err != nil {
		return err
	}
	_, err = statement.Exec(0, settings.SuccessDiscordWebhook, settings.FailureDiscordWebhook, settings.TwoCaptchaAPIKey, settings.AntiCaptchaAPIKey, settings.CapMonsterAPIKey, settings.AYCDAccessToken, settings.AYCDAPIKey, settings.DarkMode, settings.UseAnimations)

	return err
}

// AddAccount adds an Account object to the database
func AddAccount(account entities.Account) error {
	if database == nil {
		return &DatabaseNotInitializedError{}
	}

	encryptedEmail, err := common.Aes256Encrypt(account.Email, enums.UserKey)
	if err != nil {
		return err
	}

	encryptedPassword, err := common.Aes256Encrypt(account.Password, enums.UserKey)
	if err != nil {
		return err
	}

	statement, err := database.Preparex(`INSERT INTO accounts (ID, retailer, email, password, creationDate) VALUES (?, ?, ?, ?, ?)`)
	if err != nil {
		return err
	}

	_, err = statement.Exec(account.ID, account.Retailer, encryptedEmail, encryptedPassword, account.CreationDate)

	return err
}

// UpdateAccount updates an Account object in the database
func UpdateAccount(ID string, newAccount entities.Account) error {
	err := RemoveAccount(ID)
	if err != nil {
		return err
	}

	err = AddAccount(newAccount)

	return err
}

// RemoveAccount removes an Account object from the database
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
