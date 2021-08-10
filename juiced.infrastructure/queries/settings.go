package queries

import (
	"errors"
	"fmt"
	"sort"
	"time"

	"backend.juicedbot.io/juiced.infrastructure/common"
	"backend.juicedbot.io/juiced.infrastructure/common/entities"
	"backend.juicedbot.io/juiced.infrastructure/common/enums"
)

// GetSettings returns the settings object from the database
func GetSettings() (entities.Settings, error) {
	settings := entities.Settings{}
	database := common.GetDatabase()
	if database == nil {
		return settings, errors.New("database not initialized")
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
	database := common.GetDatabase()
	if database == nil {
		return accounts, errors.New("database not initialized")
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

		decryptedPassword, err := common.Aes256Decrypt(account.Password, enums.UserKey)
		if err == nil {
			account.Password = decryptedPassword
		} else {
			encryptedPassword, err := common.Aes256Encrypt(account.Password, enums.UserKey)
			if err != nil {
				return accounts, err
			}

			if encryptedPassword != "" {
				go func() {
					for {
						_, err = database.Exec(fmt.Sprintf(`UPDATE accounts SET password = "%v" WHERE ID = "%v"`, encryptedPassword, account.ID))
						if err != nil {
							continue
						} else {
							time.Sleep(1 * time.Second)
						}
					}

				}()
			}
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
	database := common.GetDatabase()
	if database == nil {
		return account, errors.New("database not initialized")
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

		decryptedPassword, err := common.Aes256Decrypt(account.Password, enums.UserKey)
		if err == nil {
			account.Password = decryptedPassword
		} else {
			encryptedPassword, err := common.Aes256Encrypt(account.Password, enums.UserKey)
			if err != nil {
				return account, err
			}

			if encryptedPassword != "" {
				go func() {
					for {
						_, err = database.Exec(fmt.Sprintf(`UPDATE accounts SET password = "%v" WHERE ID = "%v"`, encryptedPassword, account.ID))
						if err != nil {
							continue
						} else {
							time.Sleep(1 * time.Second)
						}
					}
				}()
			}
		}

	}

	return account, nil
}
