package errors

// ParseSettingsError is the error encountered when parsing JSON into a Settings object returns an error
const ParseSettingsError = "Parsing the settings returned an error: "

// GetSettingsError is the error encountered when retrieving the Settings object from the DB returns an error
const GetSettingsError = "Retrieving settings returned an error: "

// UpdateSettingsError is the error encountered when updating the Settings object from the DB returns an error
const UpdateSettingsError = "Updating settings returned an error: "

// ParseAccountError is the error encountered when parsing JSON into a Account object returns an error
const ParseAccountError = "Parsing the account returned an error: "

// AddAccountError is the error encountered when adding the Account object from the DB returns an error
const AddAccountError = "Adding the account returned an error: "

// UpdateAccountError is the error encountered when updating the Account object from the DB returns an error
const UpdateAccountError = "Updating the account returned an error: "

// RemoveAccountError is the error encountered when removing the Account object from the DB returns an error
const RemoveAccountError = "Removing the account returned an error: "

// ParseRemoveAccountsRequestError is the error encountered when parsing JSON into a DeleteAccountsRequest object returns an error
const ParseRemoveAccountsRequestError = "Parsing the JSON into a DeleteAccountsRequest returned an error: "

// TestSuccessWebhookError is the error when sending the webhook to the success webhook url fails
const TestSuccessWebhookError = "Error while sending success webhook"

// TestFailureWebhookError is the error when sending the webhook to the failure webhook url fails
const TestFailureWebhookError = "Error while sending failure webhook"
