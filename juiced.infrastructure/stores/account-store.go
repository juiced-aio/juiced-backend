package stores

import (
	"fmt"
	"sort"
	"time"

	"backend.juicedbot.io/juiced.client/http"
	"backend.juicedbot.io/juiced.infrastructure/database"
	"backend.juicedbot.io/juiced.infrastructure/entities"
	"backend.juicedbot.io/juiced.infrastructure/enums"
	"backend.juicedbot.io/juiced.infrastructure/events"
	"github.com/google/uuid"
)

type AccountStore struct {
	Accounts map[string]*entities.Account
}

var accountStore AccountStore

func InitAccountStore() error {
	accountStore = AccountStore{
		Accounts: make(map[string]*entities.Account),
	}

	accounts, err := database.GetAllAccounts()
	if err != nil {
		return err
	}

	for _, account := range accounts {
		account := account
		accountStore.Accounts[account.ID] = &account
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

	sort.SliceStable(accounts, func(i, j int) bool {
		return accounts[i].CreationDate < accounts[j].CreationDate
	})

	return accounts
}

func GetAccounts(accountIDs []string) []*entities.Account {
	accounts := []*entities.Account{}
	for _, accountID := range accountIDs {
		if account, ok := accountStore.Accounts[accountID]; ok {
			accounts = append(accounts, account)
		}
	}

	sort.SliceStable(accounts, func(i, j int) bool {
		return accounts[i].CreationDate < accounts[j].CreationDate
	})

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

func CreateTempAccount(account entities.Account) (*entities.Account, error) {
	if account.ID == "" {
		account.ID = uuid.New().String()
	}
	if account.CreationDate == 0 {
		account.CreationDate = time.Now().Unix()
	}
	account.IsTemp = true

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

func UpdateTempAccount(accountID string, newAccount entities.Account) (*entities.Account, error) {
	account, err := GetAccount(accountID)
	if err != nil {
		return nil, err
	}

	account.Email = newAccount.Email
	account.Password = newAccount.Password

	return account, nil
}

func RemoveAccount(accountID string) (entities.Account, error) {
	account, err := GetAccount(accountID)
	if err != nil {
		return entities.Account{}, err
	}

	delete(accountStore.Accounts, accountID)
	return *account, database.RemoveAccount(accountID)
}

func RemoveTempAccount(accountID string) (entities.Account, error) {
	account, err := GetAccount(accountID)
	if err != nil {
		return entities.Account{}, err
	}

	delete(accountStore.Accounts, accountID)
	return *account, nil
}

func AccessAccountCookies(accountID string) ([]*http.Cookie, error) {
	account, err := GetAccount(accountID)
	if err != nil {
		return []*http.Cookie{}, err
	}

	if len(account.Cookies) > 0 {
		if AccountIsLoggedIn(account) {
			return account.Cookies, nil
		}
	}

	AccountLogin(account)
	return []*http.Cookie{}, nil
}

func AccountLogin(account *entities.Account) error {
	if AccountIsLoggedIn(account) {
		return nil
	}

	switch account.Retailer {
	case enums.GameStop:
		go func() {
			events.GetEventBus().PublishAccountEvent(enums.AccountLoggingIn, enums.AccountStart, nil, account.ID)
			time.Sleep(5 * time.Second)
			events.GetEventBus().PublishAccountEvent(enums.AccountLoggedIn, enums.AccountComplete, nil, account.ID)
		}()
	default:
		return &enums.UnsupportedRetailerError{Retailer: account.Retailer}
	}

	return nil
}

func AccountIsLoggedIn(account *entities.Account) bool {
	// TODO
	return false
}
