package stores

import (
	"fmt"
	"time"

	"backend.juicedbot.io/juiced.infrastructure/database"
	"backend.juicedbot.io/juiced.infrastructure/entities"
	"backend.juicedbot.io/juiced.infrastructure/enums"
	"github.com/google/uuid"
)

type AccountStore struct {
	Accounts map[string]*entities.Account
}

var accountStore AccountStore

func (store *AccountStore) Init() error {
	accounts, err := database.GetAllAccounts()
	if err != nil {
		return err
	}

	for _, account := range accounts {
		store.Accounts[account.ID] = &account
	}

	return nil
}

type AccountNotFoundError struct {
	ID string
}

func (e *AccountNotFoundError) Error() string {
	return fmt.Sprintf("Account with ID %s not found", e.ID)
}

type AccountNotFoundByEmailError struct {
	Email    string
	Retailer string
}

func (e *AccountNotFoundByEmailError) Error() string {
	return fmt.Sprintf("Account with name %s not found for retailer %s", e.Email, e.Retailer)
}

func GetAllAccounts() []*entities.Account {
	accounts := []*entities.Account{}
	for _, account := range accountStore.Accounts {
		accounts = append(accounts, account)
	}

	return accounts
}

func GetAccounts(accountIDs []string) []*entities.Account {
	accounts := []*entities.Account{}
	for _, accountID := range accountIDs {
		if account, ok := accountStore.Accounts[accountID]; ok {
			accounts = append(accounts, account)
		}
	}

	return accounts
}

func GetAccount(accountID string) (*entities.Account, error) {
	profile, ok := accountStore.Accounts[accountID]
	if !ok {
		return nil, &AccountNotFoundError{accountID}
	}

	return profile, nil
}

func GetAccountByEmail(email string, retailer enums.Retailer) (*entities.Account, error) {
	for _, account := range accountStore.Accounts {
		if account.Email == email && account.Retailer == retailer {
			return account, nil
		}
	}

	return nil, &AccountNotFoundByEmailError{email, retailer}
}

func CreateAccount(account entities.Account) (*entities.Account, error) {
	if account.ID == "" {
		account.ID = uuid.New().String()
	}
	if account.CreationDate == 0 {
		account.CreationDate = time.Now().Unix()
	}

	err := database.CreateAccount(account)
	if err != nil {
		return nil, err
	}

	accountPtr := &account
	accountStore.Accounts[account.ID] = accountPtr

	return accountPtr, nil
}

func UpdateAccount(accountID string, newAccount entities.Account) (*entities.Account, error) {
	account, err := GetAccount(accountID)
	if err != nil {
		return nil, err
	}

	account.Email = newAccount.Email
	account.Password = newAccount.Password

	return account, database.UpdateAccount(accountID, *account)
}

func RemoveAccount(accountID string) (entities.Account, error) {
	account, err := GetAccount(accountID)
	if err != nil {
		return entities.Account{}, err
	}

	delete(accountStore.Accounts, accountID)
	return *account, database.RemoveAccount(accountID)
}
