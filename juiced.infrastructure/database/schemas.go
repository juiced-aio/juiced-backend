package database

var userInfoSchema = `
	CREATE TABLE IF NOT EXISTS userInfo (
		ID INTEGER,
		email TEXT,
		licenseKey TEXT,
		deviceName TEXT,
		userVer TEXT,
		discordID TEXT,
		discordUsername TEXT,
		discordAvatarURL TEXT,
		activationToken TEXT,
		refreshToken TEXT,
		expiresAt INTEGER
	)
`

var tasksSchema = `
	CREATE TABLE IF NOT EXISTS tasks (
		ID TEXT,
		taskGroupID TEXT,
		retailer TEXT,
		taskSerialized TEXT,
		creationDate INTEGER
	)
`

var taskGroupsSchema = `
	CREATE TABLE IF NOT EXISTS taskGroups (
		groupID TEXT,
		name TEXT,
		retailer TEXT,
		taskIDsSerialized TEXT,
		monitorsSerialized TEXT,
		creationDate INTEGER
	)
`

var proxyGroupsSchema = `
	CREATE TABLE IF NOT EXISTS proxyGroups (
		groupID TEXT,
		name TEXT,
		creationDate INTEGER
	)
`

var proxiesSchema = `
	CREATE TABLE IF NOT EXISTS proxies (
		ID TEXT,
		proxyGroupID TEXT,
		host TEXT,
		port TEXT,
		username TEXT,
		password TEXT
	)
`

var profileGroupsSchema = `
	CREATE TABLE IF NOT EXISTS profileGroups (
		groupID TEXT,
		name TEXT,
		profileIDsJoined TEXT,
		creationDate INTEGER
	)
`

var profilesSchema = `
	CREATE TABLE IF NOT EXISTS profiles (
		ID TEXT,
		profileGroupIDsJoined TEXT,
		name TEXT,
		email TEXT,
		phoneNumber TEXT,
		creationDate INTEGER
	)
`

var shippingAddressesSchema = `
	CREATE TABLE IF NOT EXISTS shippingAddresses (
		ID TEXT,
		profileID TEXT,
		firstName TEXT,
		lastName TEXT,
		address1 TEXT,
		address2 TEXT,
		city TEXT,
		zipCode TEXT,
		stateCode TEXT,
		countryCode TEXT
	)
`

var billingAddressesSchema = `
	CREATE TABLE IF NOT EXISTS billingAddresses (
		ID TEXT,
		profileID TEXT,
		firstName TEXT,
		lastName TEXT,
		address1 TEXT,
		address2 TEXT,
		city TEXT,
		zipCode TEXT,
		stateCode TEXT,
		countryCode TEXT
	)
`

var cardsSchema = `
	CREATE TABLE IF NOT EXISTS cards (
		ID TEXT,
		profileID TEXT,
		cardHolderName TEXT,
		cardNumber TEXT,
		expMonth TEXT,
		expYear TEXT,
		cvv TEXT,
		cardType TEXT
	)
`

var checkoutsSchema = `
	CREATE TABLE IF NOT EXISTS checkouts (
		itemName TEXT,
		imageURL TEXT,
		sku TEXT,
		price INTEGER,
		quantity INTEGER,
		retailer TEXT,
		profileName TEXT,
		msToCheckout INTEGER,
		time INTEGER
	)
`

var settingsSchema = `
	CREATE TABLE IF NOT EXISTS settings (
		id TEXT,
		successDiscordWebhook TEXT,
		failureDiscordWebhook TEXT,
		twoCaptchaAPIKey TEXT,
		antiCaptchaAPIKey TEXT,
		capMonsterAPIKey TEXT,
		aycdAccessToken TEXT,
		aycdAPIKey TEXT,
		darkMode INTEGER,
		useAnimations INTEGER
	)
`

var accountsSchema = `
	CREATE TABLE IF NOT EXISTS accounts (
		ID TEXT,
		retailer TEXT,
		email TEXT,
		password TEXT,
		cookiesSerialized TEXT,
		creationDate INTEGER
	)
`

var schemas = []string{
	userInfoSchema,

	tasksSchema,
	taskGroupsSchema,

	proxyGroupsSchema,
	proxiesSchema,

	profileGroupsSchema,
	profilesSchema,
	shippingAddressesSchema,
	billingAddressesSchema,
	cardsSchema,

	checkoutsSchema,
	settingsSchema,
	accountsSchema,
}
