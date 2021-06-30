package common

//! ALL COLUMNS MUST HAVE A DEFAULT VALUE!
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

//! ALL COLUMNS MUST HAVE A DEFAULT VALUE!
var tasksSchema = `
	CREATE TABLE IF NOT EXISTS tasks (
		ID TEXT,
		taskGroupID TEXT,
		profileID TEXT,
		proxyGroupID TEXT,
		retailer TEXT,
		sizeJoined TEXT,
		qty INTEGER,
		status TEXT,
		taskDelay INTEGER,
		creationDate INTEGER
	)
`

//! ALL COLUMNS MUST HAVE A DEFAULT VALUE!
var targetTaskInfosSchema = `
	CREATE TABLE IF NOT EXISTS targetTaskInfos (
		taskID TEXT,
		taskGroupID TEXT,
		checkoutType TEXT,
		email TEXT,
		password TEXT,
		paymentType TEXT
	)
`

//! ALL COLUMNS MUST HAVE A DEFAULT VALUE!
var walmartTaskInfosSchema = `
	CREATE TABLE IF NOT EXISTS walmartTaskInfos (
		taskID TEXT,
		taskGroupID TEXT
	)
`

//! ALL COLUMNS MUST HAVE A DEFAULT VALUE!
var amazonTaskInfosSchema = `
	CREATE TABLE IF NOT EXISTS amazonTaskInfos (
		taskID TEXT,
		taskGroupID TEXT,
		email TEXT,
		password TEXT,
		loginType TEXT
	)
`

//! ALL COLUMNS MUST HAVE A DEFAULT VALUE!
var bestbuyTaskInfosSchema = `
	CREATE TABLE IF NOT EXISTS bestbuyTaskInfos (
		taskID TEXT,
		taskGroupID TEXT,
		email TEXT,
		password TEXT,
		taskType TEXT
	)
`

//! ALL COLUMNS MUST HAVE A DEFAULT VALUE!
var gamestopTaskInfosSchema = `
	CREATE TABLE IF NOT EXISTS gamestopTaskInfos (
		taskID TEXT,
		taskGroupID TEXT,
		email TEXT,
		password TEXT,
		taskType TEXT
	)
`

//! ALL COLUMNS MUST HAVE A DEFAULT VALUE!
var taskGroupsSchema = `
	CREATE TABLE IF NOT EXISTS taskGroups (
		groupID TEXT,
		name TEXT,
		proxyGroupID TEXT,
		retailer TEXT,
		input TEXT,
		delay INTEGER,
		status TEXT,
		taskIDsJoined TEXT,
		creationDate INTEGER
	)
`

//! ALL COLUMNS MUST HAVE A DEFAULT VALUE!
var targetMonitorInfosSchema = `
	CREATE TABLE IF NOT EXISTS targetMonitorInfos (
		ID TEXT,
		taskGroupID TEXT,
		storeID TEXT,
		monitorType TEXT
	)
`

//! ALL COLUMNS MUST HAVE A DEFAULT VALUE!
var targetSingleMonitorInfosSchema = `
	CREATE TABLE IF NOT EXISTS targetSingleMonitorInfos (
		monitorID TEXT,
		taskGroupID TEXT,
		tcin TEXT,
		maxPrice INTEGER,
		checkoutType TEXT
	)
`

//! ALL COLUMNS MUST HAVE A DEFAULT VALUE!
var walmartMonitorInfosSchema = `
	CREATE TABLE IF NOT EXISTS walmartMonitorInfos (
		ID TEXT,
		taskGroupID TEXT,
		monitorType TEXT,
		skusJoined TEXT,
		maxPrice INTEGER
	)
`

//! ALL COLUMNS MUST HAVE A DEFAULT VALUE!
var amazonMonitorInfosSchema = `
	CREATE TABLE IF NOT EXISTS amazonMonitorInfos (
		ID TEXT,
		taskGroupID TEXT
	)
`

//! ALL COLUMNS MUST HAVE A DEFAULT VALUE!
var amazonSingleMonitorInfosSchema = `
	CREATE TABLE IF NOT EXISTS amazonSingleMonitorInfos (
		monitorID TEXT,
		taskGroupID TEXT,
		monitorType TEXT,
		asin TEXT,
		ofid TEXT,
		maxPrice INTEGER
	)
`

//! ALL COLUMNS MUST HAVE A DEFAULT VALUE!
var bestbuyMonitorInfosSchema = `
	CREATE TABLE IF NOT EXISTS bestbuyMonitorInfos (
		ID TEXT,
		taskGroupID TEXT
	)
`

//! ALL COLUMNS MUST HAVE A DEFAULT VALUE!
var bestbuySingleMonitorInfosSchema = `
	CREATE TABLE IF NOT EXISTS bestbuySingleMonitorInfos (
		monitorID TEXT,
		taskGroupID TEXT,
		sku TEXT,
		maxPrice INTEGER
	)
`

//! ALL COLUMNS MUST HAVE A DEFAULT VALUE!
var gamestopMonitorInfosSchema = `
	CREATE TABLE IF NOT EXISTS gamestopMonitorInfos (
		ID TEXT,
		taskGroupID TEXT
	)
`

//! ALL COLUMNS MUST HAVE A DEFAULT VALUE!
var gamestopSingleMonitorInfosSchema = `
	CREATE TABLE IF NOT EXISTS gamestopSingleMonitorInfos (
		monitorID TEXT,
		taskGroupID TEXT,
		sku TEXT,
		maxPrice INTEGER
	)
`

//! ALL COLUMNS MUST HAVE A DEFAULT VALUE!
var proxyGroupsSchema = `
	CREATE TABLE IF NOT EXISTS proxyGroups (
		groupID TEXT,
		name TEXT,
		creationDate INTEGER
	)
`

//! ALL COLUMNS MUST HAVE A DEFAULT VALUE!
var proxysSchema = `
	CREATE TABLE IF NOT EXISTS proxys (
		ID TEXT,
		proxyGroupID TEXT,
		host TEXT,
		port TEXT,
		username TEXT,
		password TEXT
	)
`

//! ALL COLUMNS MUST HAVE A DEFAULT VALUE!
var profileGroupsSchema = `
	CREATE TABLE IF NOT EXISTS profileGroups (
		groupID TEXT,
		name TEXT,
		profileIDsJoined TEXT,
		creationDate INTEGER
	)
`

//! ALL COLUMNS MUST HAVE A DEFAULT VALUE!
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

//! ALL COLUMNS MUST HAVE A DEFAULT VALUE!
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

//! ALL COLUMNS MUST HAVE A DEFAULT VALUE!
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

//! ALL COLUMNS MUST HAVE A DEFAULT VALUE!
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

//! ALL COLUMNS MUST HAVE A DEFAULT VALUE!
var checkoutsSchema = `
	CREATE TABLE IF NOT EXISTS checkouts (
		itemName TEXT,
		sku TEXT,
		price INTEGER,
		quantity INTEGER,
		retailer TEXT,
		profileName TEXT,
		msToCheckout INTEGER,
		time INTEGER
	)
`

//! ALL COLUMNS MUST HAVE A DEFAULT VALUE!
var settingsSchema = `
	CREATE TABLE IF NOT EXISTS settings (
		id TEXT,
		successDiscordWebhook TEXT DEFAULT '',
		failureDiscordWebhook TEXT DEFAULT '',
		twoCaptchaAPIKey TEXT DEFAULT '',
		antiCaptchaAPIKey TEXT DEFAULT '',
		capMonsterAPIKey TEXT DEFAULT '',
		aycdAccessToken TEXT DEFAULT '',
		aycdAPIKey TEXT DEFAULT ''
	)
`

var schemas = []string{
	userInfoSchema,
	tasksSchema,
	targetTaskInfosSchema,
	walmartTaskInfosSchema,
	amazonTaskInfosSchema,
	bestbuyTaskInfosSchema,
	gamestopTaskInfosSchema,
	taskGroupsSchema,
	targetMonitorInfosSchema,
	targetSingleMonitorInfosSchema,
	walmartMonitorInfosSchema,
	amazonMonitorInfosSchema,
	amazonSingleMonitorInfosSchema,
	bestbuyMonitorInfosSchema,
	bestbuySingleMonitorInfosSchema,
	gamestopMonitorInfosSchema,
	gamestopSingleMonitorInfosSchema,
	proxyGroupsSchema,
	proxysSchema,
	profileGroupsSchema,
	profilesSchema,
	shippingAddressesSchema,
	billingAddressesSchema,
	cardsSchema,
	checkoutsSchema,
	settingsSchema,
}
