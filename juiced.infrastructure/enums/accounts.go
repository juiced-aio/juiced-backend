package enums

type AccountStatus = string

const (
	AccountIdle               AccountStatus = "Not Logged In"
	AccountLoggingIn          AccountStatus = "Logging In"
	AccountLoggedIn           AccountStatus = "Logged In"
	AccountInvalidCredentials AccountStatus = "Invalid Credentials"
)

type AccountEventType = string

const (
	AccountStart    AccountEventType = "AccountStart"
	AccountUpdate   AccountEventType = "AccountUpdate"
	AccountFail     AccountEventType = "AccountFail"
	AccountStop     AccountEventType = "AccountStop"
	AccountComplete AccountEventType = "AccountComplete"
)
