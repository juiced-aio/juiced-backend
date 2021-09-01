package requests

type DeleteAccountsRequest struct {
	AccountIDs []string `json:"accountIDs"`
}
